package chroma

import (
	"context"

	chromatypes "github.com/amikos-tech/chroma-go/types"
	"github.com/tmc/langchaingo/embeddings"
)

var _ chromatypes.EmbeddingFunction = chromaGoEmbedder{} // compile-time check

// chromaGoEmbedder adapts an 'embeddings.Embedder' to a 'chroma_go.EmbeddingFunction'.
type chromaGoEmbedder struct {
	embeddings.Embedder
}

func (e chromaGoEmbedder) EmbedDocuments(ctx context.Context, texts []string) ([]*chromatypes.Embedding, error) {
	_embeddings, err := e.Embedder.EmbedDocuments(ctx, texts)
	if err != nil {
		return nil, err
	}
	_chrmembeddings := make([]*chromatypes.Embedding, len(_embeddings))
	for i, emb := range _embeddings {
		_chrmembeddings[i] = chromatypes.NewEmbeddingFromFloat32(emb)
	}
	return _chrmembeddings, nil
}

func (e chromaGoEmbedder) EmbedQuery(ctx context.Context, text string) (*chromatypes.Embedding, error) {
	_embedding, err := e.Embedder.EmbedQuery(ctx, text)
	if err != nil {
		return nil, err
	}
	return chromatypes.NewEmbeddingFromFloat32(_embedding), nil
}

func (e chromaGoEmbedder) EmbedRecords(ctx context.Context, records []*chromatypes.Record, force bool) error {
	return chromatypes.EmbedRecordsDefaultImpl(e, ctx, records, force)
}
