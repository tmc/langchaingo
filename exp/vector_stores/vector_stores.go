package vector_stores

import "github.com/tmc/langchaingo/schema"

// VectorStore is the interface for saving and querying documents in the form of vector embeddings.
type VectorStore interface {
	AddDocuments(documents []schema.Document) error
	SimilaritySearch(query string, numDocuments int) ([]schema.Document, error)
}
