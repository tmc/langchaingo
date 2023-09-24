package ernie

import (
	"context"

	"github.com/tmc/langchaingo/embeddings"
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

// split texts with batchCount.
func (e *Ernie) embed(ctx context.Context, texts []string) ([][]float32, error) {
	emb := make([][]float32, 0, len(texts))

	offsetLen := len(texts) / e.batchCount
	for i := 0; i <= offsetLen; i++ {
		start := i * e.batchCount
		end := i*e.batchCount + e.batchCount

		if end > len(texts) {
			end = len(texts)
		}

		curTextEmbeddings, err := e.client.CreateEmbedding(ctx, texts[start:end])
		if err != nil {
			return nil, err
		}

		emb = append(emb, curTextEmbeddings...)
	}
	return emb, nil
}

// EmbedDocuments use ernie Embedding-V1.
func (e *Ernie) EmbedDocuments(ctx context.Context, texts []string) ([][]float32, error) {
	batchedTexts := embeddings.BatchTexts(
		embeddings.MaybeRemoveNewLines(texts, e.stripNewLines),
		e.batchSize,
	)

	emb := make([][]float32, 0, len(texts))
	for _, texts := range batchedTexts {
		curTextEmbeddings, err := e.embed(ctx, texts)
		if err != nil {
			return nil, err
		}

		textLengths := make([]int, 0, len(texts))
		for _, text := range texts {
			textLengths = append(textLengths, len(text))
		}

		combined, err := embeddings.CombineVectors(curTextEmbeddings, textLengths)
		if err != nil {
			return nil, err
		}

		emb = append(emb, combined)
	}

	return emb, nil
}

// EmbedQuery use ernie Embedding-V1.
func (e *Ernie) EmbedQuery(ctx context.Context, text string) ([]float64, error) {
	emb, err := e.EmbedDocuments(ctx, []string{text})
	if err != nil {
		return nil, err
	}

	return emb[0], nil
}
