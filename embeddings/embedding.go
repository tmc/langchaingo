package embeddings

import (
	"context"
	"strings"
)

// NewEmbedder creates a new Embedder from the given EmbedderClient, with
// some options that affect how embedding will be done.
func NewEmbedder(client EmbedderClient, opts ...Option) (*EmbedderImpl, error) {
	e := &EmbedderImpl{
		client:        client,
		StripNewLines: defaultStripNewLines,
		BatchSize:     defaultBatchSize,
	}

	for _, opt := range opts {
		opt(e)
	}
	return e, nil
}

// Embedder is the interface for creating vector embeddings from texts.
type Embedder interface {
	// EmbedDocuments returns a vector for each text.
	EmbedDocuments(ctx context.Context, texts []string) ([][]float32, error)
	// EmbedQuery embeds a single text.
	EmbedQuery(ctx context.Context, text string) ([]float32, error)
}

// EmbedderClient is the interface LLM clients implement for embeddings.
type EmbedderClient interface {
	CreateEmbedding(ctx context.Context, texts []string) ([][]float32, error)
}

type EmbedderImpl struct {
	client EmbedderClient

	StripNewLines bool
	BatchSize     int
}

// EmbedQuery embeds a single text.
func (ei *EmbedderImpl) EmbedQuery(ctx context.Context, text string) ([]float32, error) {
	if ei.StripNewLines {
		text = strings.ReplaceAll(text, "\n", " ")
	}

	emb, err := ei.client.CreateEmbedding(ctx, []string{text})
	if err != nil {
		return nil, err
	}

	return emb[0], nil
}

// EmbedDocuments creates one vector embedding for each of the texts.
func (ei *EmbedderImpl) EmbedDocuments(ctx context.Context, texts []string) ([][]float32, error) {
	texts = MaybeRemoveNewLines(texts, ei.StripNewLines)
	return BatchedEmbed(ctx, ei.client, texts, ei.BatchSize)
}

func MaybeRemoveNewLines(texts []string, removeNewLines bool) []string {
	if !removeNewLines {
		return texts
	}

	for i := 0; i < len(texts); i++ {
		texts[i] = strings.ReplaceAll(texts[i], "\n", " ")
	}

	return texts
}

// BatchTexts splits strings by the length batchSize.
func BatchTexts(texts []string, batchSize int) [][]string {
	batchedTexts := make([][]string, len(texts))
	for i, text := range texts {
		runeText := []rune(text)

		for j := 0; j < len(runeText); j += batchSize {
			if j+batchSize >= len(runeText) {
				batchedTexts[i] = append(batchedTexts[i], string(runeText[j:]))
				break
			}

			batchedTexts[i] = append(batchedTexts[i], string(runeText[j:j+batchSize]))
		}
	}

	return batchedTexts
}

// BatchedEmbed creates embeddings for the given input texts, batching them
// into batches of batchSize if needed.
func BatchedEmbed(ctx context.Context, embedder EmbedderClient, texts []string, batchSize int) ([][]float32, error) {
	batchedTexts := BatchTexts(texts, batchSize)

	emb := make([][]float32, 0, len(texts))
	for _, batch := range batchedTexts {
		curBatchEmbeddings, err := embedder.CreateEmbedding(ctx, batch)
		if err != nil {
			return nil, err
		}
		emb = append(emb, curBatchEmbeddings...)
	}

	return emb, nil
}
