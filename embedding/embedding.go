package embedding

import "context"

// Embedder is the interface for creating vector embeddings from texts.
type Embedder interface {
	// EmbedDocuments returns a vector for each text.
	EmbedDocuments(ctx context.Context, texts []string) ([][]float64, error)
	// EmbedQuery embeds a single text.
	EmbedQuery(ctx context.Context, text string) ([]float64, error)
}
