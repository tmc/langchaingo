package chroma

import (
	"context"
	"errors"
	"fmt"

	chromago "github.com/amikos-tech/chroma-go"
	"github.com/amikos-tech/chroma-go/openai"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/vectorstores"
)

var (
	ErrInvalidScoreThreshold    = errors.New("score threshold must be between 0 and 1")
	ErrUnexpectedResponseLength = errors.New("unexpected length of response")
	ErrResetDB                  = errors.New("error resetting database")
	ErrNewClient                = errors.New("error creating collection")
	ErrRemoveCollection         = errors.New("error resetting collection")
	ErrUnsupportedOptions       = errors.New("unsupported options")
)

// Store is a wrapper around the chromaGo API and client.
type Store struct {
	client           *chromago.Client
	collection       *chromago.Collection
	distanceFunction chromago.DistanceFunction
	resetChroma      bool
	chromaURL        string
	openaiAPIKey     string
	collectionName   string
	// TODO (noodnik2): clarify need for / support of the following fields
	// nameSpace   string
	// embedder    embeddings.Embedder
	// grpcConn    *grpc.ClientConn
	// client      pinecone_grpc.VectorServiceClient
	// indexName   string
	// projectName string
	// environment string
	// apiKey      string
	// textKey     string
	// useGRPC     bool
}

// TODO (noodnik2): (why) is this needed?
// var _ vectorstores.VectorStore = Store{}

func New(_ context.Context, opts ...Option) (Store, error) {
	s, coErr := applyClientOptions(opts...)
	if coErr != nil {
		return s, coErr
	}

	s.client = chromago.NewClient(s.chromaURL)
	if s.resetChroma {
		// TODO (noodnik2): is this really needed?
		if _, errRest := s.client.Reset(); errRest != nil {
			return s, fmt.Errorf("%w: %w", ErrResetDB, errRest)
		}
	}

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
	if opts.NameSpace != "" || opts.Embedder != nil || opts.ScoreThreshold != 0 || opts.Filters != nil {
		// TODO (noodnik2): implement (some of?) these
		return ErrUnsupportedOptions
	}

	ids := make([]string, len(docs))
	texts := make([]string, len(docs))
	metadatas := make([]map[string]any, len(docs))
	for i, doc := range docs {
		texts[i] = doc.PageContent
		metadatas[i] = doc.Metadata
		ids[i] = fmt.Sprintf("id%d", i+1) // TODO (noodnik2): clarify meaning / use of "ids"
	}

	col := s.collection
	if _, addErr := col.Add(nil, metadatas, texts, ids); addErr != nil {
		return fmt.Errorf("adding documents: %w", addErr)
	}
	return nil
}

func (s Store) SimilaritySearch(_ context.Context, query string, numDocuments int,
	options ...vectorstores.Option,
) ([]schema.Document, error) {
	opts := s.getOptions(options...)

	if opts.NameSpace != "" || opts.Embedder != nil {
		// TODO (noodnik2): implement these
		return nil, fmt.Errorf("%w: NameSpace, Embedder", ErrUnsupportedOptions)
	}

	scoreThreshold, stErr := s.getScoreThreshold(opts)
	if stErr != nil {
		return nil, stErr
	}
	filters, _ := s.getFilters(opts).(map[string]any)
	qr, queryErr := s.collection.Query([]string{query}, int32(numDocuments), filters, nil, nil)
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

// TODO (noodnik2): does this map to chroma.Store.collectionName?  E.g., for consistency with
//  the existing model, or to leave it in the local "chroma.Option" where it is now?
//  func (s Store) getNameSpace(opts vectorstores.Options) string {
//  	if opts.NameSpace != "" {
//  		return opts.NameSpace
//  	}
//  	return s.nameSpace
//  }

func (s Store) getScoreThreshold(opts vectorstores.Options) (float64, error) {
	if opts.ScoreThreshold < 0 || opts.ScoreThreshold > 1 {
		return 0, ErrInvalidScoreThreshold
	}
	return opts.ScoreThreshold, nil
}

func (s Store) getFilters(opts vectorstores.Options) any {
	if opts.Filters != nil {
		return opts.Filters
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
