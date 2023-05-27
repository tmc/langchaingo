package embeddings

import (
	"context"
	"strings"

	"github.com/tmc/langchaingo/llms/vertexai"
)

// VertexAIPaLM is the embedder using the Google PaLM api.
type VertexAIPaLM struct {
	client *vertexai.LLM

	StripNewLines bool
	BatchSize     int
}

var _ Embedder = VertexAIPaLM{}

// NewVertexAIPaLM creates a new VertexAI with StripNewLines set to true and batch
// size set to 512.
func NewVertexAIPaLM() (*VertexAIPaLM, error) {
	client, err := vertexai.New()
	if err != nil {
		return nil, err
	}
	return &VertexAIPaLM{
		client:        client,
		StripNewLines: true,
		BatchSize:     defaultBatchSize,
	}, nil
}

// EmbedDocuments creates one vector embedding for each of the texts.
func (e VertexAIPaLM) EmbedDocuments(ctx context.Context, texts []string) ([][]float64, error) {
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
func (e VertexAIPaLM) EmbedQuery(ctx context.Context, text string) ([]float64, error) {
	if e.StripNewLines {
		text = strings.ReplaceAll(text, "\n", " ")
	}

	embeddings, err := e.client.CreateEmbedding(ctx, []string{text})
	if err != nil {
		return nil, err
	}

	return embeddings[0], nil
}
