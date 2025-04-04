package cybertron

import (
	"context"

	"github.com/averikitsch/langchaingo/embeddings"
	"github.com/nlpodyssey/cybertron/pkg/models/bert"
	"github.com/nlpodyssey/cybertron/pkg/tasks/textencoding"
)

// Cybertron is the embedder using Cybertron to run embedding models locally.
type Cybertron struct {
	encoder         textencoding.Interface
	Model           string
	ModelsDir       string
	PoolingStrategy bert.PoolingStrategyType
}

var _ embeddings.EmbedderClient = (*Cybertron)(nil)

// NewCybertron returns a new embedding client that uses Cybertron to run embedding
// models locally (on the CPU). The embedding model will be downloaded and cached
// automatically. Use `WithModel` and `WithModelsDir` to change which model is used
// and where it is cached.
func NewCybertron(opts ...Option) (*Cybertron, error) {
	return applyOptions(opts...)
}

// CreateEmbedding implements the `embeddings.EmbedderClient` and creates an embedding
// vector for each of the supplied texts.
func (c *Cybertron) CreateEmbedding(ctx context.Context, texts []string) ([][]float32, error) {
	result := make([][]float32, 0, len(texts))

	for _, text := range texts {
		embedding, err := c.encoder.Encode(ctx, text, int(c.PoolingStrategy))
		if err != nil {
			return nil, err
		}

		result = append(result, embedding.Vector.Normalize2().Data().F32())
	}

	return result, nil
}
