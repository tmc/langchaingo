package vertexai

import (
	"context"
	"strings"

	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/llms/vertexai"
)

// VertexAIPaLM is the embedder using the Google PaLM api.
type VertexAIPaLM struct { //nolint:revive
	client *vertexai.LLM

	StripNewLines bool
	BatchSize     int
}

var _ embeddings.Embedder = VertexAIPaLM{}

// NewVertexAIPaLM creates a new VertexAI with options. Options for client, strip new lines and batch size.
func NewVertexAIPaLM(opts ...Option) (*VertexAIPaLM, error) {
	v, err := applyClientOptions(opts...)
	if err != nil {
		return nil, err
	}

	return v, nil
}

// EmbedDocuments creates one vector embedding for each of the texts.
func (e VertexAIPaLM) EmbedDocuments(ctx context.Context, texts []string) ([][]float64, error) {
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
func (e VertexAIPaLM) EmbedQuery(ctx context.Context, text string) ([]float64, error) {
	if e.StripNewLines {
		text = strings.ReplaceAll(text, "\n", " ")
	}

	emb, err := e.client.CreateEmbedding(ctx, []string{text})
	if err != nil {
		return nil, err
	}

	return emb[0], nil
}
