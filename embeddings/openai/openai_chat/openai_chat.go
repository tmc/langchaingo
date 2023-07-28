package openai_chat

import (
	"context"
	"strings"

	"github.com/tmc/langchaingo/embeddings"
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

func (e ChatOpenAI) EmbedDocuments(ctx context.Context, texts []string) ([][]float64, error) {
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

func (e ChatOpenAI) EmbedQuery(ctx context.Context, text string) ([]float64, error) {
	if e.StripNewLines {
		text = strings.ReplaceAll(text, "\n", " ")
	}

	emb, err := e.client.CreateEmbedding(ctx, []string{text})
	if err != nil {
		return nil, err
	}

	return emb[0], nil
}
