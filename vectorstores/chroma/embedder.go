package chroma

import (
	"context"

	chromaembed "github.com/amikos-tech/chroma-go/pkg/embeddings"
	"github.com/tmc/langchaingo/embeddings"
)

var _ chromaembed.EmbeddingFunction = chromaGoEmbedder{} // compile-time check

type chromaGoEmbedder struct {
	embeddings.Embedder
}

func (e chromaGoEmbedder) EmbedDocuments(ctx context.Context, texts []string) ([]chromaembed.Embedding, error) {
	vecs, err := e.Embedder.EmbedDocuments(ctx, texts)
	if err != nil {
		return nil, err
	}
	out := make([]chromaembed.Embedding, len(vecs))
	for i, v := range vecs {
		out[i] = chromaembed.NewEmbeddingFromFloat32(v)
	}
	return out, nil
}

func (e chromaGoEmbedder) EmbedQuery(ctx context.Context, text string) (chromaembed.Embedding, error) {
	v, err := e.Embedder.EmbedQuery(ctx, text)
	if err != nil {
		return nil, err
	}
	return chromaembed.NewEmbeddingFromFloat32(v), nil
}

func (e chromaGoEmbedder) Name() string { return "langchaingo" }

func (e chromaGoEmbedder) GetConfig() chromaembed.EmbeddingFunctionConfig { return nil }

func (e chromaGoEmbedder) DefaultSpace() chromaembed.DistanceMetric { return chromaembed.L2 }

func (e chromaGoEmbedder) SupportedSpaces() []chromaembed.DistanceMetric {
	return []chromaembed.DistanceMetric{chromaembed.L2, chromaembed.COSINE, chromaembed.IP}
}
