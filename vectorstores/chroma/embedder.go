package chroma

import (
	"context"

	chroma_go "github.com/amikos-tech/chroma-go"
	"github.com/tmc/langchaingo/embeddings"
)

var _ chroma_go.EmbeddingFunction = chromaGoEmbedder{} // compile-time check

// chromaGoEmbedder adapts an 'embeddings.Embedder' to a 'chroma_go.EmbeddingFunction'.
type chromaGoEmbedder struct {
	embeddings.Embedder
}

// CreateEmbeddingWithModel passes through to `CreateEmbedding`; i.e., ignores 'model' (second) parameter.
func (e chromaGoEmbedder) CreateEmbeddingWithModel(texts []string, _ string) ([][]float32, error) {
	return e.CreateEmbedding(texts)
}

// CreateEmbedding creates the embedding using the current model.
func (e chromaGoEmbedder) CreateEmbedding(texts []string) ([][]float32, error) {
	return e.EmbedDocuments(context.TODO(), texts)
}
