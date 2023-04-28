package embedding

import (
	"context"
	"strings"

	"github.com/tmc/langchaingo/llms/openai"
)

type OpenAI struct {
	client        *openai.LLM
	StripNewLines bool
	BatchSize     int
}

var _ Embedder = OpenAI{}

func NewOpenAI() (OpenAI, error) {
	client, err := openai.New()
	if err != nil {
		return OpenAI{}, err
	}

	return OpenAI{
		client:        client,
		StripNewLines: true,
		BatchSize:     512,
	}, nil
}

// EmbedDocuments creates a vector embedding for each of the texts.
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
		combined, err := combineVectors(curTextEmbeddings)

		embeddings = append(embeddings, combined)
	}

	return embeddings, nil
}

func (e OpenAI) EmbedQuery(ctx context.Context, text string) ([]float64, error) {
	if e.StripNewLines {
		text = strings.ReplaceAll(text, "\n", " ")
	}

	embeddings, err := e.client.CreateEmbedding(ctx, []string{text})
	if err != nil {
		return []float64{}, err
	}

	return embeddings[0], nil
}

func maybeRemoveNewLines(texts []string, removeNewLines bool) []string {
	if !removeNewLines {
		return texts
	}

	for i := 0; i < len(texts); i++ {
		texts[i] = strings.ReplaceAll(texts[i], "\n", " ")
	}

	return texts
}

// batchTexts splits strings by the length batchSize.
func batchTexts(texts []string, batchSize int) [][]string {
	batchedTexts := make([][]string, len(texts))
	for i, text := range texts {
		runeText := []rune(text)

		for j := 0; j < len(runeText); j += batchSize {
			if j+batchSize >= len(runeText) {
				batchedTexts[i] = append(batchedTexts[i], string(runeText[j:]))
				break
			}

			batchedTexts[i] = append(batchedTexts[i], string(runeText[j:j+batchSize]))
		}
	}

	return batchedTexts
}

func combineVectors(vectors [][]float64) ([]float64, error) {
	return vectors[0], nil
}
