package ernie

import (
	"context"

	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/llms/ernie"
)

// Ernie https://cloud.baidu.com/doc/WENXINWORKSHOP/s/alj562vvu
type Ernie struct {
	client        *ernie.LLM
	batchSize     int
	stripNewLines bool
}

var _ embeddings.Embedder = &Ernie{}

// NewErnie creates a new Ernie with options. Options for client, strip new lines and batch size.
func NewErnie(opts ...Option) (*Ernie, error) {
	v := &Ernie{
		stripNewLines: defaultStripNewLines,
		batchSize:     defaultBatchSize,
	}

	for _, opt := range opts {
		opt(v)
	}

	if v.client == nil {
		client, err := ernie.New()
		if err != nil {
			return nil, err
		}
		v.client = client
	}

	return v, nil
}

// EmbedDocuments implements embeddings.Embedder .
// simple impl.
func (e *Ernie) EmbedDocuments(ctx context.Context, texts []string) ([][]float64, error) {
	batchedTexts := embeddings.BatchTexts(
		embeddings.MaybeRemoveNewLines(texts, e.stripNewLines),
		e.batchSize,
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
