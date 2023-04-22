package embeddings

// Embeddings is the interface for creating vector embeddings from texts.
type Embeddings interface {
	// Returns vector for each text
	EmbedDocuments(texts []string) ([][]float64, error)
	// Embeds a single text
	EmbedQuery(text string) ([]float64, error)
}
