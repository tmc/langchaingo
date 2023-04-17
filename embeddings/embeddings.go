package embeddings

type Embeddings interface {
	EmbedDocuments(texts []string) ([][]float64, error)
	EmbedQuery(text string) ([]float64, error)
}
