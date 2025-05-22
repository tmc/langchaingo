package chroma

import (
	"context"

	chromaembedding "github.com/amikos-tech/chroma-go/pkg/embeddings"
	"github.com/tmc/langchaingo/embeddings"
)

var _ chromaembedding.EmbeddingFunction = chromaGoEmbedderV2{} // compile-time check

// chromaGoEmbedder adapts an 'embeddings.Embedder' to a 'chroma_go.EmbeddingFunction'.
type chromaGoEmbedderV2 struct {
	embeddings.Embedder
}

func (e chromaGoEmbedderV2) EmbedDocuments(ctx context.Context, texts []string) ([]chromaembedding.Embedding, error) {
	_embeddings, err := e.Embedder.EmbedDocuments(ctx, texts)
	if err != nil {
		return nil, err
	}
	_chrmembeddings := make([]chromaembedding.Embedding, len(_embeddings))
	for i, emb := range _embeddings {
		_chrmembeddings[i] = chromaembedding.NewEmbeddingFromFloat32(emb)
	}
	return _chrmembeddings, nil
}

func (e chromaGoEmbedderV2) EmbedQuery(ctx context.Context, text string) (chromaembedding.Embedding, error) {
	_embedding, err := e.Embedder.EmbedQuery(ctx, text)
	if err != nil {
		return nil, err
	}
	return chromaembedding.NewEmbeddingFromFloat32(_embedding), nil
}
