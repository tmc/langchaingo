package schema

import "context"

// Retriever is an interface that defines the behavior of a retriever.
type Retriever interface {
	GetRelevantDocuments(ctx context.Context, query string) ([]Document, error)
}
