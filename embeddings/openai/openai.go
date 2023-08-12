package openai

import (
	"context"
	"strings"

	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/llms/openai"
)

// OpenAI is the embedder using the OpenAI api.
type OpenAI struct {
	client *openai.LLM

	StripNewLines bool
	BatchSize     int
}

var _ embeddings.Embedder = OpenAI{}

// NewOpenAI creates a new OpenAI with options. Options for client, strip new lines and batch.
func NewOpenAI(opts ...Option) (OpenAI, error) {
	o, err := applyClientOptions(opts...)
	if err != nil {
		return OpenAI{}, err
	}

	return o, nil
}

// EmbedDocuments creates one vector embedding for each of the texts.
func (e OpenAI) EmbedDocuments(ctx context.Context, texts []string) ([][]float64, error) {
	batchedTexts := embeddings.BatchTexts(
		embeddings.MaybeRemoveNewLines(texts, e.StripNewLines),
		e.BatchSize,
	)

	emb := make([][]float64, 0, len(texts))
	for _, texts := range batchedTexts {
		curTextEmbeddings, err := e.client.CreateEmbedding(ctx, texts)
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

// EmbedQuery embeds a single text.
func (e OpenAI) EmbedQuery(ctx context.Context, text string) ([]float64, error) {
	if e.StripNewLines {
		text = strings.ReplaceAll(text, "\n", " ")
	}

	emb, err := e.client.CreateEmbedding(ctx, []string{text})
	if err != nil {
		return nil, err
	}

	return emb[0], nil
}
