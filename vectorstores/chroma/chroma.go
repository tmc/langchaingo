package chroma

import (
	"context"
	"errors"
	"fmt"

	chromago "github.com/amikos-tech/chroma-go"
	"github.com/amikos-tech/chroma-go/openai"
	chromatypes "github.com/amikos-tech/chroma-go/types"
	"github.com/google/uuid"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/vectorstores"
	"golang.org/x/exp/maps"
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
	client             *chromago.Client
	collection         *chromago.Collection
	distanceFunction   chromatypes.DistanceFunction
	chromaURL          string
	openaiAPIKey       string
	openaiOrganization string

	nameSpace    string
	nameSpaceKey string
	embedder     embeddings.Embedder
	includes     []chromatypes.QueryEnum
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
	chromaClient, err := chromago.NewClient(s.chromaURL)
	if err != nil {
		return s, err
	}

	if _, errHb := chromaClient.Heartbeat(context.Background()); errHb != nil {
		return s, errHb
	}
	s.client = chromaClient

	var embeddingFunction chromatypes.EmbeddingFunction
	if s.embedder != nil {
		// inject user's embedding function, if provided
		embeddingFunction = chromaGoEmbedder{Embedder: s.embedder}
	} else {
		// otherwise use standard langchaingo OpenAI embedding function
		var options []openai.Option
		if s.openaiOrganization != "" {
			options = append(options, openai.WithOpenAIOrganizationID(s.openaiOrganization))
		}
		embeddingFunction, err = openai.NewOpenAIEmbeddingFunction(s.openaiAPIKey, options...)
		if err != nil {
			return s, err
		}
	}

	col, errCc := s.client.CreateCollection(context.Background(), s.nameSpace, map[string]any{}, true,
		embeddingFunction, s.distanceFunction)
	if errCc != nil {
		return s, fmt.Errorf("%w: %w", ErrNewClient, errCc)
	}

	s.collection = col
	return s, nil
}

// AddDocuments adds the text and metadata from the documents to the Chroma collection associated with 'Store'.
// and returns the ids of the added documents.
func (s Store) AddDocuments(ctx context.Context,
	docs []schema.Document,
	options ...vectorstores.Option,
) ([]string, error) {
	opts := s.getOptions(options...)
	if opts.Embedder != nil || opts.ScoreThreshold != 0 || opts.Filters != nil {
		return nil, ErrUnsupportedOptions
	}

	nameSpace := s.getNameSpace(opts)
	if nameSpace != "" && s.nameSpaceKey == "" {
		return nil, fmt.Errorf("%w: nameSpace without nameSpaceKey", ErrUnsupportedOptions)
	}

	ids := make([]string, len(docs))
	texts := make([]string, len(docs))
	metadatas := make([]map[string]any, len(docs))
	for docIdx, doc := range docs {
		ids[docIdx] = uuid.New().String() // TODO (noodnik2): find & use something more meaningful
		texts[docIdx] = doc.PageContent
		mc := make(map[string]any, 0)
		maps.Copy(mc, doc.Metadata)
		metadatas[docIdx] = mc
		if nameSpace != "" {
			metadatas[docIdx][s.nameSpaceKey] = nameSpace
		}
	}

	col := s.collection
	if _, addErr := col.Add(ctx, nil, metadatas, texts, ids); addErr != nil {
		return nil, fmt.Errorf("%w: %w", ErrAddDocument, addErr)
	}
	return ids, nil
}

func (s Store) SimilaritySearch(ctx context.Context, query string, numDocuments int,
	options ...vectorstores.Option,
) ([]schema.Document, error) {
	opts := s.getOptions(options...)

	if opts.Embedder != nil {
		// embedder is not used by this method, so shouldn't ever be specified
		return nil, fmt.Errorf("%w: Embedder", ErrUnsupportedOptions)
	}

	scoreThreshold, stErr := s.getScoreThreshold(opts)
	if stErr != nil {
		return nil, stErr
	}

	filter := s.getNamespacedFilter(opts)
	qr, queryErr := s.collection.Query(ctx, []string{query}, safeIntToInt32(numDocuments), filter, nil, s.includes)
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
			if score := 1.0 - qr.Distances[docsI][docI]; score >= scoreThreshold {
				sDocs = append(sDocs, schema.Document{
					Metadata:    qr.Metadatas[docsI][docI],
					PageContent: qr.Documents[docsI][docI],
					Score:       score,
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
	_, errDc := s.client.DeleteCollection(context.Background(), s.collection.Name)
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

func (s Store) getScoreThreshold(opts vectorstores.Options) (float32, error) {
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

func safeIntToInt32(n int) int32 {
	return int32(max(0, n))
}
