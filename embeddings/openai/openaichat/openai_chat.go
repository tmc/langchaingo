package openaichat

import (
	"context"
	"strings"

	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/embeddings/internal/embedderclient"
	"github.com/tmc/langchaingo/llms/openai"
)

// ChatOpenAI is the embedder using the OpenAI api.
type ChatOpenAI struct {
	client *openai.Chat

	StripNewLines bool
	BatchSize     int
}

var _ embeddings.Embedder = ChatOpenAI{}

// NewChatOpenAI creates a new ChatOpenAI with options. Options for client, strip new lines and batch.
func NewChatOpenAI(opts ...ChatOption) (ChatOpenAI, error) {
	o, err := applyChatClientOptions(opts...)
	if err != nil {
		return ChatOpenAI{}, err
	}

	return o, nil
}

func (e ChatOpenAI) EmbedDocuments(ctx context.Context, texts []string) ([][]float32, error) {
	texts = embeddings.MaybeRemoveNewLines(texts, e.StripNewLines)
	return embedderclient.BatchedEmbed(ctx, e.client, texts, e.BatchSize)
}

func (e ChatOpenAI) EmbedQuery(ctx context.Context, text string) ([]float32, error) {
	if e.StripNewLines {
		text = strings.ReplaceAll(text, "\n", " ")
	}

	emb, err := e.client.CreateEmbedding(ctx, []string{text})
	if err != nil {
		return nil, err
	}

	return emb[0], nil
}
