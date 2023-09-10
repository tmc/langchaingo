package chroma

import (
	"context"
	"errors"
	"fmt"

	chromago "github.com/amikos-tech/chroma-go"
	"github.com/amikos-tech/chroma-go/openai"
	"github.com/tmc/langchaingo/internal/util"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/vectorstores"
)

var (
	ErrInvalidScoreThreshold    = errors.New("score threshold must be between 0 and 1")
	ErrUnexpectedResponseLength = errors.New("unexpected length of response")
	ErrNewClient                = errors.New("error creating collection")
	ErrAddDocument              = errors.New("error adding document")
	ErrRemoveCollection         = errors.New("error resetting collection")
	ErrUnsupportedOptions       = errors.New("unsupported options")
)

// Store is a wrapper around the chromaGo API and client.
type Store struct {
	client           *chromago.Client
	collection       *chromago.Collection
	distanceFunction chromago.DistanceFunction
	chromaURL        string
	openaiAPIKey     string
	collectionName   string
	nameSpace        string
	nameSpaceKey     string
	// embedder    embeddings.Embedder  // TODO (noodnik2): implement embedder consistent with other adapters
	// indexName   string // TODO (noodnik2): Chroma equivalent?  https://docs.pinecone.io/docs/indexes
	// projectName string // TODO (noodnik2): Chroma equivalent?  https://docs.pinecone.io/docs/projects
	// textKey     string // TODO (noodnik2): Is this called for / needed?
}

var _ vectorstores.VectorStore = Store{}

// New creates an active client connection to the (specified, or default) collection in the Chroma server
// and returns the `Store` object needed by the other accessors.
func New(opts ...Option) (Store, error) {
	s, coErr := applyClientOptions(opts...)
	if coErr != nil {
		return s, coErr
	}

	// create the client connection and confirm that we can access the server with it
	chromaClient := chromago.NewClient(s.chromaURL)
	if _, errHb := chromaClient.Heartbeat(); errHb != nil {
		return s, errHb
	}
	s.client = chromaClient

	embeddingFunction := openai.NewOpenAIEmbeddingFunction(s.openaiAPIKey)
	// TODO (noodnik2): integrate "embedding function" similar to the other vectorstore adapters
	col, errCc := s.client.CreateCollection(s.collectionName, map[string]any{}, true,
		embeddingFunction, s.distanceFunction)
	if errCc != nil {
		return s, fmt.Errorf("%w: %w", ErrNewClient, errCc)
	}

	s.collection = col
	return s, nil
}

func (s Store) AddDocuments(_ context.Context, docs []schema.Document, options ...vectorstores.Option) error {
	opts := s.getOptions(options...)
	if opts.Embedder != nil || opts.ScoreThreshold != 0 || opts.Filters != nil {
		return ErrUnsupportedOptions
	}

	nameSpace := s.getNameSpace(opts)
	if nameSpace != "" && s.nameSpaceKey == "" {
		return fmt.Errorf("%w: nameSpace without nameSpaceKey", ErrUnsupportedOptions)
	}

	ids := make([]string, len(docs))
	texts := make([]string, len(docs))
	metadatas := make([]map[string]any, len(docs))
	for docIdx, doc := range docs {
		// TODO (noodnik2): making up an "id" value here seems meaningless; is
		//  there a "well-known" metadata (or other) value we can use instead?
		ids[docIdx] = fmt.Sprintf("%s-%d", nameSpace, docIdx)
		texts[docIdx] = doc.PageContent
		metadatas[docIdx] = util.NewMap(doc.Metadata)
		if nameSpace != "" {
			metadatas[docIdx][s.nameSpaceKey] = nameSpace
		}
	}

	col := s.collection
	if _, addErr := col.Add(nil, metadatas, texts, ids); addErr != nil {
		return fmt.Errorf("%w: %w", ErrAddDocument, addErr)
	}
	return nil
}

func (s Store) SimilaritySearch(_ context.Context, query string, numDocuments int,
	options ...vectorstores.Option,
) ([]schema.Document, error) {
	opts := s.getOptions(options...)

	if opts.Embedder != nil {
		// TODO (noodnik2): implement these
		return nil, fmt.Errorf("%w: Embedder", ErrUnsupportedOptions)
	}

	scoreThreshold, stErr := s.getScoreThreshold(opts)
	if stErr != nil {
		return nil, stErr
	}

	filter := s.getNamespacedFilter(opts)
	qr, queryErr := s.collection.Query([]string{query}, int32(numDocuments), filter, nil, nil)
	if queryErr != nil {
		return nil, queryErr
	}

	if len(qr.Documents) != len(qr.Metadatas) || len(qr.Metadatas) != len(qr.Distances) {
		return nil, fmt.Errorf("%w: qr.Documents[%d], qr.Metadatas[%d], qr.Distances[%d]",
			ErrUnexpectedResponseLength, len(qr.Documents), len(qr.Metadatas), len(qr.Distances))
	}
	var sDocs []schema.Document
	for docsI := range qr.Documents {
		for docI := range qr.Documents[docsI] {
			distanceFound := float64(qr.Distances[docsI][docI])
			if (1.0 - distanceFound) >= scoreThreshold {
				sDocs = append(sDocs, schema.Document{
					Metadata:    qr.Metadatas[docsI][docI],
					PageContent: qr.Documents[docsI][docI],
				})
			}
		}
	}

	return sDocs, nil
}

func (s Store) RemoveCollection() error {
	if s.client == nil || s.collection == nil {
		return fmt.Errorf("%w: no collection", ErrRemoveCollection)
	}
	_, errDc := s.client.DeleteCollection(s.collection.Name)
	if errDc != nil {
		return fmt.Errorf("%w(%s): %w", ErrRemoveCollection, s.collection.Name, errDc)
	}
	return nil
}

func (s Store) getOptions(options ...vectorstores.Option) vectorstores.Options {
	opts := vectorstores.Options{}
	for _, opt := range options {
		opt(&opts)
	}
	return opts
}

func (s Store) getScoreThreshold(opts vectorstores.Options) (float64, error) {
	if opts.ScoreThreshold < 0 || opts.ScoreThreshold > 1 {
		return 0, ErrInvalidScoreThreshold
	}
	return opts.ScoreThreshold, nil
}

func (s Store) getNameSpace(opts vectorstores.Options) string {
	if opts.NameSpace != "" {
		return opts.NameSpace
	}
	return s.nameSpace
}

func (s Store) getNamespacedFilter(opts vectorstores.Options) map[string]any {
	filter, _ := opts.Filters.(map[string]any)

	nameSpace := s.getNameSpace(opts)
	if nameSpace == "" || s.nameSpaceKey == "" {
		return filter
	}

	nameSpaceFilter := map[string]any{s.nameSpaceKey: nameSpace}
	if filter == nil {
		return nameSpaceFilter
	}

	return map[string]any{"$and": []map[string]any{nameSpaceFilter, filter}}
}
