package vertex

import (
	"context"
	"errors"
	"fmt"

	"google.golang.org/genai"
)

// CreateEmbedding creates embeddings from texts using the new google.golang.org/genai library.
func (g *Vertex) CreateEmbedding(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, errors.New("no texts provided")
	}

	// Convert text strings to genai.Content format
	contents := make([]*genai.Content, len(texts))
	for i, text := range texts {
		contents[i] = genai.NewContentFromText(text, genai.RoleUser)
	}

	// Get the embedding model name from options
	modelName := g.opts.DefaultEmbeddingModel
	if modelName == "" {
		modelName = "text-embedding-004" // Default Gemini embedding model
	}

	// Use the new EmbedContent API
	resp, err := g.client.Models.EmbedContent(ctx, modelName, contents, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to embed content: %w", err)
	}

	if resp.Embeddings == nil || len(resp.Embeddings) == 0 {
		return nil, errors.New("empty embedding response")
	}

	// Convert response to [][]float32 format
	// resp.Embeddings is a slice of ContentEmbedding
	embeddings := make([][]float32, len(resp.Embeddings))
	for i, emb := range resp.Embeddings {
		if emb.Values == nil || len(emb.Values) == 0 {
			return nil, fmt.Errorf("empty embedding at index %d", i)
		}
		embeddings[i] = make([]float32, len(emb.Values))
		copy(embeddings[i], emb.Values)
	}

	if len(texts) != len(embeddings) {
		return embeddings, fmt.Errorf("returned %d embeddings for %d texts", len(embeddings), len(texts))
	}

	return embeddings, nil
}
