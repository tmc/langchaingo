package chroma

import (
	"context"

	chroma "github.com/amikos-tech/chroma-go"
	"github.com/tmc/langchaingo/embeddings"
)

var _ chroma.EmbeddingFunction = chromaGoEmbedder{}

// chromaGoEmbedder adapts a langchaingo embedder to a chroma_go.EmbeddingFunction.
type chromaGoEmbedder struct {
	embeddings.Embedder
}

func (e chromaGoEmbedder) CreateEmbeddingWithModel(texts []string, _ string) ([][]float32, error) {
	// passthru
	return e.CreateEmbedding(texts)
}

func (e chromaGoEmbedder) CreateEmbedding(texts []string) ([][]float32, error) {
	// get the [][]float64 embeddings from langchaingo's embedder
	vectors, err := e.EmbedDocuments(context.TODO(), texts)
	if err != nil {
		return nil, err
	}
	// convert them to the [][]float32 shape returned from chromago's embedder
	target := make([][]float32, len(vectors))
	for row := 0; row < len(vectors); row++ {
		target[row] = make([]float32, len(vectors[row]))
		for col := 0; col < len(vectors[row]); col++ {
			target[row][col] = float32(vectors[row][col])
		}
	}
	return target, nil
}
