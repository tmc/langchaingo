package pinecone

import (
	"context"
	"errors"

	"github.com/pinecone-io/go-pinecone/pinecone_grpc"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/vectorstores"
	"google.golang.org/grpc"
)

var (
	// ErrMissingTextKey is returned in SimilaritySearch if a vector
	// from the query is missing the text key.
	ErrMissingTextKey = errors.New("missing text key in vector metadata")
	// ErrEmbedderWrongNumberVectors is returned when if the embedder returns a number
	// of vectors that is not equal to the number of documents given.
	ErrEmbedderWrongNumberVectors = errors.New(
		"number of vectors from embedder does not match number of documents",
	)
	// ErrEmptyResponse is returned if the API gives an empty response.
	ErrEmptyResponse         = errors.New("empty response")
	ErrInvalidScoreThreshold = errors.New(
		"score threshold must be between 0 and 1")
)

// Store is a wrapper around the pinecone rest API and grpc client.
type Store struct {
	embedder embeddings.Embedder
	grpcConn *grpc.ClientConn
	client   pinecone_grpc.VectorServiceClient

	indexName   string
	projectName string
	environment string
	apiKey      string
	textKey     string
	nameSpace   string
	useGRPC     bool
}

var _ vectorstores.VectorStore = Store{}

// New creates a new Store with options. Options for index name, environment, project name
// and embedder must be set.
func New(ctx context.Context, opts ...Option) (Store, error) {
	s, err := applyClientOptions(opts...)
	if err != nil {
		return Store{}, err
	}

	if s.useGRPC {
		conn, err := s.getGRPCConn(ctx)
		if err != nil {
			return Store{}, err
		}
		s.grpcConn = conn

		client := pinecone_grpc.NewVectorServiceClient(conn)
		s.client = client
	}

	return s, nil
}

// AddDocuments creates vector embeddings from the documents using the embedder
// and upsert the vectors to the pinecone index.
func (s Store) AddDocuments(ctx context.Context, docs []schema.Document, options ...vectorstores.Option) error {
	opts := s.getOptions(options...)

	nameSpace := s.getNameSpace(opts)

	texts := make([]string, 0, len(docs))
	for _, doc := range docs {
		texts = append(texts, doc.PageContent)
	}

	vectors, err := s.embedder.EmbedDocuments(ctx, texts)
	if err != nil {
		return err
	}

	if len(vectors) != len(docs) {
		return ErrEmbedderWrongNumberVectors
	}

	metadatas := make([]map[string]any, 0, len(docs))
	for i := 0; i < len(docs); i++ {
		metadata := make(map[string]any, len(docs[i].Metadata))
		for key, value := range docs[i].Metadata {
			metadata[key] = value
		}
		metadata[s.textKey] = texts[i]

		metadatas = append(metadatas, metadata)
	}

	if s.useGRPC {
		return s.grpcUpsert(ctx, vectors, metadatas, nameSpace)
	}

	return s.restUpsert(ctx, vectors, metadatas, nameSpace)
}

// SimilaritySearch creates a vector embedding from the query using the embedder
// and queries to find the most similar documents.
func (s Store) SimilaritySearch(ctx context.Context, query string, numDocuments int, options ...vectorstores.Option) ([]schema.Document, error) { //nolint:lll
	opts := s.getOptions(options...)

	nameSpace := s.getNameSpace(opts)

	filters := s.getFilters(opts)

	scoreThreshold, err := s.getScoreThreshold(opts)
	if err != nil {
		return nil, err
	}

	vector, err := s.embedder.EmbedQuery(ctx, query)
	if err != nil {
		return nil, err
	}

	if s.useGRPC {
		return s.grpcQuery(ctx, vector, numDocuments, nameSpace)
	}

	return s.restQuery(ctx, vector, numDocuments, nameSpace, scoreThreshold,
		filters)
}

// Close closes the grpc connection.
func (s Store) Close() error {
	return s.grpcConn.Close()
}

func (s Store) getNameSpace(opts vectorstores.Options) string {
	if opts.NameSpace != "" {
		return opts.NameSpace
	}
	return s.nameSpace
}

func (s Store) getScoreThreshold(opts vectorstores.Options) (float64,
	error,
) {
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
