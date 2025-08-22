package vertex

import (
	"context"
	"errors"
	"fmt"

	aiplatform "cloud.google.com/go/aiplatform/apiv1"
	"cloud.google.com/go/aiplatform/apiv1/aiplatformpb"
	"google.golang.org/api/option"
	"google.golang.org/protobuf/types/known/structpb"
)

// CreateEmbedding creates embeddings from texts using Gemini embedding models.
// This replaces the deprecated PaLM embedding API with the new Gemini embedding API.
// The default model is "text-embedding-004" but can be configured via DefaultEmbeddingModel option.
func (v *Vertex) CreateEmbedding(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, errors.New("no texts provided")
	}

	// Use the newer aiplatform client for embeddings since the genai client doesn't support them
	apiEndpoint := fmt.Sprintf("%s-aiplatform.googleapis.com:443", v.opts.CloudLocation)

	client, err := aiplatform.NewPredictionClient(ctx, option.WithEndpoint(apiEndpoint))
	if err != nil {
		return nil, fmt.Errorf("failed to create prediction client: %w", err)
	}
	defer client.Close()

	// Use configured embedding model or default to text-embedding-004
	model := v.opts.DefaultEmbeddingModel
	if model == "" {
		model = "text-embedding-004"
	}

	endpoint := fmt.Sprintf("projects/%s/locations/%s/publishers/google/models/%s",
		v.opts.CloudProject, v.opts.CloudLocation, model)

	// Prepare instances for batch embedding
	instances := make([]*structpb.Value, 0, len(texts))
	for _, text := range texts {
		instance, err := structpb.NewStruct(map[string]interface{}{
			"content": text,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create instance: %w", err)
		}
		instances = append(instances, structpb.NewStructValue(instance))
	}

	// Prepare parameters (optional - can specify output dimensionality)
	params, err := structpb.NewStruct(map[string]interface{}{})
	if err != nil {
		return nil, fmt.Errorf("failed to create parameters: %w", err)
	}

	req := &aiplatformpb.PredictRequest{
		Endpoint:   endpoint,
		Instances:  instances,
		Parameters: structpb.NewStructValue(params),
	}

	resp, err := client.Predict(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get embeddings: %w", err)
	}

	// Extract embeddings from response
	results := make([][]float32, 0, len(texts))
	for _, prediction := range resp.GetPredictions() {
		embeddingStruct := prediction.GetStructValue()
		if embeddingStruct == nil {
			return nil, errors.New("invalid embedding response structure")
		}

		// Get the embeddings array from the response
		embeddingField := embeddingStruct.GetFields()["embeddings"]
		if embeddingField == nil {
			return nil, errors.New("embeddings field not found in response")
		}

		// Get the values array
		valuesField := embeddingField.GetStructValue().GetFields()["values"]
		if valuesField == nil {
			return nil, errors.New("values field not found in embeddings")
		}

		// Convert to float32 array
		values := valuesField.GetListValue().GetValues()
		embedding := make([]float32, len(values))
		for i, v := range values {
			embedding[i] = float32(v.GetNumberValue())
		}
		results = append(results, embedding)
	}

	if len(results) != len(texts) {
		return results, fmt.Errorf("returned %d embeddings for %d texts", len(results), len(texts))
	}

	return results, nil
}
