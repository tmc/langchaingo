package chromadb

import (
	"context"

	chroma "github.com/amikos-tech/chroma-go"
	"github.com/tmc/langchaingo/embeddings"
)

var _ chroma.EmbeddingFunction = wrappedEmbeddingFunction{}

// wrappedEmbeddingFunction is a wrapper around an embeddings.
type wrappedEmbeddingFunction struct {
	embeddings.Embedder
}

func (embedder wrappedEmbeddingFunction) CreateEmbedding(documents []string) ([][]float32, error) {
	vectors, err := embedder.EmbedDocuments(context.TODO(), documents)
	if err != nil {
		return nil, err
	}
	target := make([][]float32, len(vectors))
	for row := 0; row < len(vectors); row++ {
		target[row] = make([]float32, len(vectors[row]))
		for col := 0; col < len(vectors[row]); col++ {
			target[row][col] = float32(vectors[row][col])
		}
	}
	return target, nil
}

func (embedder wrappedEmbeddingFunction) CreateEmbeddingWithModel(documents []string, model string) ([][]float32, error) {
	return embedder.CreateEmbedding(documents)
}
