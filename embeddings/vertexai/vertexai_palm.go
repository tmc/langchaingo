package vertexai

import (
	"context"
	"strings"

	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/embeddings/internal/embedderclient"
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
func (e VertexAIPaLM) EmbedDocuments(ctx context.Context, texts []string) ([][]float32, error) {
	texts = embeddings.MaybeRemoveNewLines(texts, e.StripNewLines)
	return embedderclient.BatchedEmbed(ctx, e.client, texts, e.BatchSize)
}

// EmbedQuery embeds a single text.
func (e VertexAIPaLM) EmbedQuery(ctx context.Context, text string) ([]float32, error) {
	if e.StripNewLines {
		text = strings.ReplaceAll(text, "\n", " ")
	}

	emb, err := e.client.CreateEmbedding(ctx, []string{text})
	if err != nil {
		return nil, err
	}

	return emb[0], nil
}
