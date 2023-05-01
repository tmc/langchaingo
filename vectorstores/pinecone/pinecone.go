package pinecone

import (
	"context"
	"errors"

	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/vectorstores"
	"github.com/tmc/langchaingo/vectorstores/pinecone/internal/grpcapi"
	"github.com/tmc/langchaingo/vectorstores/pinecone/internal/restapi"
)

var (
	// ErrMissingTextKey is returned in SimilaritySearch if a vector
	// from the query is missing the text key.
	ErrMissingTextKey             = errors.New("missing text key in vector metadata")
	ErrEmbedderWrongNumberVectors = errors.New(
		"number of vectors from embedder does not match number of documents",
	)
)

// Store is a wrapper around the pinecone rest API and grpc client.
type Store struct {
	embedder embeddings.Embedder

	indexName   string
	projectName string
	environment string
	apiKey      string
	textKey     string
	nameSpace   string
	useGRPC     bool
}

var _ vectorstores.VectorStore = Store{}

// New crates a new Store with options.
func New(opts ...Option) (Store, error) {
	return applyClientOptions(opts...)
}

// AddDocuments creates vector embeddings from the documents using the embedder
// and upsert the vectors to the pinecone index.
func (s Store) AddDocuments(ctx context.Context, docs []schema.Document) error {
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
		conn, err := grpcapi.GetGRPCConn(ctx, s.indexName, s.projectName, s.environment, s.apiKey)
		if err != nil {
			return err
		}

		return grpcapi.Upsert(ctx, conn, vectors, metadatas, s.nameSpace)
	}

	return restapi.Upsert(
		ctx,
		vectors,
		metadatas,
		s.apiKey,
		s.nameSpace,
		s.indexName,
		s.projectName,
		s.environment,
	)
}

// SimilaritySearch creates a vector embedding from the query using the embedder
// and queries to find the most similar documents.
func (s Store) SimilaritySearch(ctx context.Context, query string, numDocuments int) ([]schema.Document, error) {
	vector, err := s.embedder.EmbedQuery(ctx, query)
	if err != nil {
		return nil, err
	}

	if s.useGRPC {
		conn, err := grpcapi.GetGRPCConn(ctx, s.indexName, s.projectName, s.environment, s.apiKey)
		if err != nil {
			return nil, err
		}

		docs, err := grpcapi.Query(
			ctx,
			conn,
			vector,
			numDocuments,
			s.nameSpace,
			s.textKey,
		)

		if err != nil && errors.Is(err, grpcapi.ErrMissingTextKey) {
			return nil, ErrMissingTextKey
		}

		return docs, err
	}

	docs, err := restapi.Query(
		ctx,
		vector,
		numDocuments,
		s.apiKey,
		s.textKey,
		s.nameSpace,
		s.indexName,
		s.projectName,
		s.environment,
	)

	if err != nil && errors.Is(err, restapi.ErrMissingTextKey) {
		return nil, ErrMissingTextKey
	}

	return docs, err
}

func (s Store) ToRetriever(numDocuments int) Retriever {
	return Retriever{
		s:            s,
		numDocuments: numDocuments,
	}
}

type Retriever struct {
	s            Store
	numDocuments int
}

var _ schema.Retriever = Retriever{}

func (r Retriever) GetRelevantDocuments(ctx context.Context, query string) ([]schema.Document, error) {
	return r.s.SimilaritySearch(ctx, query, r.numDocuments)
}
