package vectorstores

import (
	"context"

	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/schema"
)

// VectorStore is the interface for saving and querying documents in the
// form of vector embeddings.
type VectorStore interface {
	AddDocuments(ctx context.Context, docs []schema.Document, options ...Option) ([]string, error)
	SimilaritySearch(ctx context.Context, query string, numDocuments int, options ...Option) ([]schema.Document, error) //nolint:lll
}

// Retriever is a retriever for vector stores.
type Retriever struct {
	CallbacksHandler callbacks.Handler
	v                VectorStore
	numDocs          int
	options          []Option
}

var _ schema.Retriever = Retriever{}

// GetRelevantDocuments returns documents using the vector store.
func (r Retriever) GetRelevantDocuments(ctx context.Context, query string) ([]schema.Document, error) {
	if r.CallbacksHandler != nil {
		r.CallbacksHandler.HandleRetrieverStart(ctx, query)
	}

	docs, err := r.v.SimilaritySearch(ctx, query, r.numDocs, r.options...)
	if err != nil {
		return nil, err
	}

	if r.CallbacksHandler != nil {
		r.CallbacksHandler.HandleRetrieverEnd(ctx, query, docs)
	}

	return docs, nil
}

// ToRetriever takes a vector store and returns a retriever using the
// vector store to retrieve documents.
func ToRetriever(vectorStore VectorStore, numDocuments int, options ...Option) Retriever {
	return Retriever{
		v:       vectorStore,
		numDocs: numDocuments,
		options: options,
	}
}
