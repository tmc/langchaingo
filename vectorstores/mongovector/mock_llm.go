package mongovector

import (
	"context"

	"github.com/vendasta/langchaingo/embeddings"
)

// mockLLM will create consistent text embeddings mocking the OpenAI
// text-embedding-3-small algorithm.
type mockLLM struct {
	seen map[string][]float32
	dim  int
}

var _ embeddings.EmbedderClient = &mockLLM{}

// createEmbedding will return vector embeddings for the mock LLM, maintaining
// consistency.
func (emb *mockLLM) CreateEmbedding(_ context.Context, texts []string) ([][]float32, error) {
	if emb.seen == nil {
		emb.seen = map[string][]float32{}
	}

	vectors := make([][]float32, len(texts))
	for i, text := range texts {
		if f32s := emb.seen[text]; len(f32s) > 0 {
			vectors[i] = f32s

			continue
		}

		vectors[i] = newNormalizedVector(emb.dim)
		emb.seen[text] = vectors[i] // ensure consistency
	}

	return vectors, nil
}
