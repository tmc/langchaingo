package schema

// Retriever is an interface that defines the behavior of a retriever.
type Retriever interface {
	GetRelevantDocuments(query string) ([]Document, error)
}
