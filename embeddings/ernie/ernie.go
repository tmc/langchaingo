package ernie

import (
	"context"

	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/llms/ernie"
)

// Ernie https://cloud.baidu.com/doc/WENXINWORKSHOP/s/alj562vvu
type Ernie struct {
	client *ernie.LLM
}

var _ embeddings.Embedder = &Ernie{}

// todo: use option pass, more: https://cloud.baidu.com/doc/WENXINWORKSHOP/s/alj562vvu#body%E5%8F%82%E6%95%B0
const batchSize = 16

func NewErnie() (*Ernie, error) {
	llm, e := ernie.New()
	if e != nil {
		return nil, e
	}
	return &Ernie{client: llm}, nil
}

// EmbedDocuments implements embeddings.Embedder .
// simple impl.
func (e *Ernie) EmbedDocuments(ctx context.Context, texts []string) ([][]float64, error) {
	batchedTexts := embeddings.BatchTexts(
		embeddings.MaybeRemoveNewLines(texts, true),
		batchSize,
	)

	emb := make([][]float64, 0, len(texts))
	for _, curTexts := range batchedTexts {
		curTextEmbeddings, err := e.client.CreateEmbedding(ctx, curTexts)
		if err != nil {
			return nil, err
		}
		emb = append(emb, curTextEmbeddings...)
	}

	return emb, nil
}

// EmbedQuery implements embeddings.Embedder.
func (e *Ernie) EmbedQuery(ctx context.Context, text string) ([]float64, error) {
	emb, err := e.client.CreateEmbedding(ctx, []string{text})
	if err != nil {
		return nil, err
	}

	return emb[0], nil
}
