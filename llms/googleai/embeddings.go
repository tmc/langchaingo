package googleai

import (
	"context"
	"fmt"
)

// CreateEmbedding creates embeddings from texts.
// TODO: Implement embeddings using new SDK API
// The new SDK's embedding API structure needs to be confirmed
func (g *GoogleAI) CreateEmbedding(ctx context.Context, texts []string) ([][]float32, error) {
	return nil, fmt.Errorf("embeddings API not yet implemented for new SDK - please use the old SDK for embedding features")
}
