package embedderclient

import (
	"context"

	"github.com/tmc/langchaingo/embeddings"
)

// EmbedderClient is the interface LLM clients implement for embeddings.
type EmbedderClient interface {
	CreateEmbedding(ctx context.Context, texts []string) ([][]float32, error)
}

// BatchedEmbed creates embeddings for the given input texts, batching them
// into batches of batchSize if needed.
func BatchedEmbed(ctx context.Context, embedder EmbedderClient, texts []string, batchSize int) ([][]float32, error) {
	batchedTexts := embeddings.BatchTexts(texts, batchSize)

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

		combined, err := embeddings.CombineVectors(curTextEmbeddings, textLengths)
		if err != nil {
			return nil, err
		}

		emb = append(emb, combined)
	}

	return emb, nil
}
