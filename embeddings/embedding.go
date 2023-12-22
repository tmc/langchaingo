package embeddings

import (
	"context"
	"strings"

	"github.com/tmc/langchaingo/internal/util"
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
	batchedTexts := make([][]string, 0, len(texts)/batchSize+1)

	for i := 0; i < len(texts); i += batchSize {
		batchedTexts = append(batchedTexts, texts[i:util.MinInt([]int{i + batchSize, len(texts)})])
	}

	return batchedTexts
}

// BatchedEmbed creates embeddings for the given input texts, batching them
// into batches of batchSize if needed.
func BatchedEmbed(ctx context.Context, embedder EmbedderClient, texts []string, batchSize int) ([][]float32, error) {
	batchedTexts := BatchTexts(texts, batchSize)

	emb := make([][]float32, 0, len(texts))
	for _, texts := range batchedTexts {
		curTextEmbeddings, err := embedder.CreateEmbedding(ctx, texts)
		if err != nil {
			return nil, err
		}
		// If the size of this batch is 1, don't average/combine the vectors.
		if len(texts) == 1 {
			emb = append(emb, curTextEmbeddings[0])
			continue
		}

		textLengths := make([]int, 0, len(texts))
		for _, text := range texts {
			textLengths = append(textLengths, len(text))
		}

		combined, err := CombineVectors(curTextEmbeddings, textLengths)
		if err != nil {
			return nil, err
		}

		emb = append(emb, combined)
	}

	return emb, nil
}
