package mistral

import (
	"context"
	"errors"
)

var ErrEmptyEmbeddings = errors.New("empty embeddings")

const defaultEmbeddingModel = "mistral-embed"

func convertFloat64ToFloat32(input []float64) []float32 {
	// Create a slice with the same length as the input.
	output := make([]float32, len(input))

	// Iterate over the input slice and convert each element.
	for i, v := range input {
		output[i] = float32(v)
	}

	return output
}

// CreateEmbedding implements the embeddings.EmbedderClient interface and creates embeddings for the given input texts.
func (m *Model) CreateEmbedding(_ context.Context, inputTexts []string) ([][]float32, error) {
	model := defaultEmbeddingModel
	if m.clientOptions.modelFromCaller {
		model = m.clientOptions.model
	}
	embsRes, err := m.client.Embeddings(model, inputTexts)
	if err != nil {
		return nil, errors.New("failed to create embeddings: " + err.Error())
	}
	allEmbds := make([][]float32, len(embsRes.Data))
	for i, embs := range embsRes.Data {
		if len(embs.Embedding) == 0 {
			return nil, ErrEmptyEmbeddings
		}
		allEmbds[i] = convertFloat64ToFloat32(embs.Embedding)
	}
	return allEmbds, nil
}
