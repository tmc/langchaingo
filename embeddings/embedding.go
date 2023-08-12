package embeddings

import (
	"context"
	"strings"
)

// Embedder is the interface for creating vector embeddings from texts.
type Embedder interface {
	// EmbedDocuments returns a vector for each text.
	EmbedDocuments(ctx context.Context, texts []string) ([][]float64, error)
	// EmbedQuery embeds a single text.
	EmbedQuery(ctx context.Context, text string) ([]float64, error)
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
