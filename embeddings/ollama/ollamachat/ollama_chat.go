package ollamachat

import (
	"context"
	"strings"

	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/embeddings/internal/embedderclient"
	"github.com/tmc/langchaingo/llms/ollama"
)

// ChatOllama is the embedder using the Ollama api.
type ChatOllama struct {
	client *ollama.Chat

	StripNewLines bool
	BatchSize     int
}

var _ embeddings.Embedder = ChatOllama{}

// NewChatOllama creates a new ChatOllama with options. Options for client, strip new lines and batch.
func NewChatOllama(opts ...ChatOption) (ChatOllama, error) {
	o, err := applyChatClientOptions(opts...)
	if err != nil {
		return ChatOllama{}, err
	}

	return o, nil
}

func (e ChatOllama) EmbedDocuments(ctx context.Context, texts []string) ([][]float32, error) {
	texts = embeddings.MaybeRemoveNewLines(texts, e.StripNewLines)
	return embedderclient.BatchedEmbed(ctx, e.client, texts, e.BatchSize)
}

func (e ChatOllama) EmbedQuery(ctx context.Context, text string) ([]float32, error) {
	if e.StripNewLines {
		text = strings.ReplaceAll(text, "\n", " ")
	}

	emb, err := e.client.CreateEmbedding(ctx, []string{text})
	if err != nil {
		return nil, err
	}

	return emb[0], nil
}
