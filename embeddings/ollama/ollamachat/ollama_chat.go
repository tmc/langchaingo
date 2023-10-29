package ollamachat

import (
	"context"
	"strings"

	"github.com/tmc/langchaingo/embeddings"
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
	batchedTexts := embeddings.BatchTexts(
		embeddings.MaybeRemoveNewLines(texts, e.StripNewLines),
		e.BatchSize,
	)

	emb := make([][]float32, 0, len(texts))
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
