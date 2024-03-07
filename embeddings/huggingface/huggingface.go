package huggingface

import (
	"context"
	"strings"

	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/llms/huggingface"
)

// Huggingface is the embedder using the Huggingface hub api.
type Huggingface struct {
	client *huggingface.LLM
	Model  string
	Task   string

	StripNewLines bool
	BatchSize     int
}

var _ embeddings.Embedder = &Huggingface{}

func NewHuggingface(opts ...Option) (*Huggingface, error) {
	v, err := applyOptions(opts...)
	if err != nil {
		return nil, err
	}

	return v, nil
}

func (e *Huggingface) EmbedDocuments(ctx context.Context, texts []string) ([][]float32, error) {
	batchedTexts := embeddings.BatchTexts(
		embeddings.MaybeRemoveNewLines(texts, e.StripNewLines),
		e.BatchSize,
	)

	emb := make([][]float32, 0, len(texts))
	for _, batch := range batchedTexts {
		curBatchEmbeddings, err := e.client.CreateEmbedding(ctx, batch, e.Model, e.Task)
		if err != nil {
			return nil, err
		}
		emb = append(emb, curBatchEmbeddings...)
	}

	return emb, nil
}

func (e *Huggingface) EmbedQuery(ctx context.Context, text string) ([]float32, error) {
	if e.StripNewLines {
		text = strings.ReplaceAll(text, "\n", " ")
	}

	emb, err := e.client.CreateEmbedding(ctx, []string{text}, e.Model, e.Task)
	if err != nil {
		return nil, err
	}

	return emb[0], nil
}
