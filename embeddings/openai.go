package embeddings

import (
	"context"
	"strings"

	"github.com/tmc/langchaingo/llms/openai"
)

// defaultBatchSize is the default length of batches.
const defaultBatchSize = 512

// OpenAI is the embedder using the OpenAI api.
type OpenAI struct {
	client *openai.LLM

	StripNewLines bool
	BatchSize     int
}

var _ Embedder = OpenAI{}

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
	batchedTexts := batchTexts(
		maybeRemoveNewLines(texts, e.StripNewLines),
		e.BatchSize,
	)

	embeddings := make([][]float64, 0, len(texts))
	for _, texts := range batchedTexts {
		curTextEmbeddings, err := e.client.CreateEmbedding(ctx, texts)
		if err != nil {
			return nil, err
		}

		textLengths := make([]int, 0, len(texts))
		for _, text := range texts {
			textLengths = append(textLengths, len(text))
		}

		combined, err := combineVectors(curTextEmbeddings, textLengths)
		if err != nil {
			return nil, err
		}

		embeddings = append(embeddings, combined)
	}

	return embeddings, nil
}

// EmbedQuery embeds a single text.
func (e OpenAI) EmbedQuery(ctx context.Context, text string) ([]float64, error) {
	if e.StripNewLines {
		text = strings.ReplaceAll(text, "\n", " ")
	}

	embeddings, err := e.client.CreateEmbedding(ctx, []string{text})
	if err != nil {
		return nil, err
	}

	return embeddings[0], nil
}
