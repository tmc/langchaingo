package ernie

import (
	"context"

	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/embeddings/internal/embedderclient"
	"github.com/tmc/langchaingo/llms/ernie"
)

// Ernie Embedding-V1 doc: https://cloud.baidu.com/doc/WENXINWORKSHOP/s/alj562vvu
type Ernie struct {
	client        *ernie.LLM
	batchSize     int // 每个文本长度不超过 384个token
	batchCount    int // 文本数量不超过16
	stripNewLines bool
}

var _ embeddings.Embedder = &Ernie{}

// NewErnie creates a new Ernie with options. Options for client, strip new lines and batch size.
func NewErnie(opts ...Option) (*Ernie, error) {
	v := &Ernie{
		stripNewLines: defaultStripNewLines,
		batchSize:     defaultBatchSize,
		batchCount:    defaultBatchCount,
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

// EmbedDocuments use ernie Embedding-V1.
func (e *Ernie) EmbedDocuments(ctx context.Context, texts []string) ([][]float32, error) {
	texts = embeddings.MaybeRemoveNewLines(texts, e.stripNewLines)
	return embedderclient.BatchedEmbed(ctx, e.client, texts, e.batchSize)
}

// EmbedQuery use ernie Embedding-V1.
func (e *Ernie) EmbedQuery(ctx context.Context, text string) ([]float32, error) {
	emb, err := e.EmbedDocuments(ctx, []string{text})
	if err != nil {
		return nil, err
	}

	return emb[0], nil
}
