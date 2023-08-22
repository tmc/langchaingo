package vertexaichat

import (
	"context"
	"strings"

	"github.com/tmc/langchaingo/embeddings"
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

func (e ChatVertexAI) EmbedDocuments(ctx context.Context, texts []string) ([][]float64, error) {
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

func (e ChatVertexAI) EmbedQuery(ctx context.Context, text string) ([]float64, error) {
	if e.StripNewLines {
		text = strings.ReplaceAll(text, "\n", " ")
	}

	emb, err := e.client.CreateEmbedding(ctx, []string{text})
	if err != nil {
		return nil, err
	}

	return emb[0], nil
}
