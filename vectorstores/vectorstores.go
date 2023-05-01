package vectorstores

import (
	"context"

	"github.com/tmc/langchaingo/schema"
)

// VectorStore is the interface for saving and querying documents in the form of vector embeddings.
type VectorStore interface {
	AddDocuments(context.Context, []schema.Document) error
	SimilaritySearch(ctx context.Context, query string, numDocuments int) ([]schema.Document, error)
}
