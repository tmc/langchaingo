package vectorstores

import (
	"context"

	"github.com/tmc/langchaingo/schema"
)

// VectorStore is the interface for saving and querying documents in the
// form of vector embeddings.
type VectorStore interface {
	AddDocuments(context.Context, []schema.Document, ...Option) error
	SimilaritySearch(ctx context.Context, query string, numDocuments int, options ...Option) ([]schema.Document, error) //nolint:lll
}

// Retriever is a retriever for vector stores.
type Retriever struct {
	v       VectorStore
	numDocs int
}

var _ schema.Retriever = Retriever{}

// GetRelevantDocuments returns documents using the vector store.
func (r Retriever) GetRelevantDocuments(ctx context.Context, query string) ([]schema.Document, error) {
	return r.v.SimilaritySearch(ctx, query, r.numDocs)
}

// ToRetriever takes a vector store and returns a retriever using the
// vector store to retrieve documents.
func ToRetriever(vectorStore VectorStore, numDocuments int) Retriever {
	return Retriever{
		v:       vectorStore,
		numDocs: numDocuments,
	}
}
