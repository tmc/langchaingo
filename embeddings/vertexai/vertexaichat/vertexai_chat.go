package vertexaichat

import (
	"context"
	"strings"

	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/embeddings/internal/embedderclient"
	"github.com/tmc/langchaingo/llms/vertexai"
)

// ChatVertexAI is the embedder using the VertexAI api.
type ChatVertexAI struct {
	client *vertexai.Chat

	StripNewLines bool
	BatchSize     int
}

var _ embeddings.Embedder = ChatVertexAI{}

// NewChatVertexAI creates a new ChatVertexAI with options. Options for client, strip new lines and batch.
func NewChatVertexAI(opts ...ChatOption) (ChatVertexAI, error) {
	o, err := applyChatClientOptions(opts...)
	if err != nil {
		return ChatVertexAI{}, err
	}

	return o, nil
}

func (e ChatVertexAI) EmbedDocuments(ctx context.Context, texts []string) ([][]float32, error) {
	texts = embeddings.MaybeRemoveNewLines(texts, e.StripNewLines)
	return embedderclient.BatchedEmbed(ctx, e.client, texts, e.BatchSize)
}

func (e ChatVertexAI) EmbedQuery(ctx context.Context, text string) ([]float32, error) {
	if e.StripNewLines {
		text = strings.ReplaceAll(text, "\n", " ")
	}

	emb, err := e.client.CreateEmbedding(ctx, []string{text})
	if err != nil {
		return nil, err
	}

	return emb[0], nil
}
