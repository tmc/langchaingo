package googleai

import (
	"context"

	"github.com/google/generative-ai-go/genai"
)

// CreateEmbedding creates embeddings from texts.
func (g *GoogleAI) CreateEmbedding(ctx context.Context, texts []string) ([][]float32, error) {
	em := g.client.EmbeddingModel(g.opts.DefaultEmbeddingModel)

	results := make([][]float32, 0, len(texts))
	for _, t := range texts {
		res, err := em.EmbedContent(ctx, genai.Text(t))
		if err != nil {
			return results, err
		}
		results = append(results, res.Embedding.Values)
	}

	return results, nil
}
