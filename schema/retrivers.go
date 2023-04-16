package schema

type Retriever interface {
	GetRelevantDocuments(query string) ([]Document, error)
}
