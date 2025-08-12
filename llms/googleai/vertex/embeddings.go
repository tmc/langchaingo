package vertex

import (
	"context"
	"errors"
	"fmt"

	"github.com/tmc/langchaingo/llms/googleai/internal/palmclient"
)

// CreateEmbedding creates embeddings from texts.
func (g *Vertex) CreateEmbedding(ctx context.Context, texts []string) ([][]float32, error) {
	embeddings, err := g.palmClient.CreateEmbedding(ctx, &palmclient.EmbeddingRequest{
		Input:     texts,
		Dimension: g.opts.DefaultEmbeddingDimension,
	})
	if err != nil {
		return [][]float32{}, err
	}

	if len(embeddings) == 0 {
		return nil, errors.New("empty response")
	}
	if len(texts) != len(embeddings) {
		return embeddings, fmt.Errorf("returned %d embeddings for %d texts", len(embeddings), len(texts))
	}

	return embeddings, nil
}
