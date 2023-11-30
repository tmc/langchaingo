package ollama

import (
	"context"
	"strings"

	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/llms/ollama"
)

// Ollama is the embedder using the Ollama api.
type Ollama struct {
	client *ollama.LLM

	StripNewLines bool
	BatchSize     int
}

var _ embeddings.Embedder = Ollama{}

// NewOllama creates a new Ollama with options. Options for client, strip new lines and batch.
func NewOllama(opts ...Option) (Ollama, error) {
	o, err := applyClientOptions(opts...)
	if err != nil {
		return Ollama{}, err
	}

	return o, nil
}

// EmbedDocuments creates one vector embedding for each of the texts.
func (e Ollama) EmbedDocuments(ctx context.Context, texts []string) ([][]float32, error) {
	texts = embeddings.MaybeRemoveNewLines(texts, e.StripNewLines)
	return embeddings.BatchedEmbed(ctx, e.client, texts, e.BatchSize)
}

// EmbedQuery embeds a single text.
func (e Ollama) EmbedQuery(ctx context.Context, text string) ([]float32, error) {
	if e.StripNewLines {
		text = strings.ReplaceAll(text, "\n", " ")
	}

	emb, err := e.client.CreateEmbedding(ctx, []string{text})
	if err != nil {
		return nil, err
	}

	return emb[0], nil
}
