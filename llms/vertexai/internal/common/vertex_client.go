package common

import (
	"context"
	"fmt"
	"runtime"
	"strings"

	"github.com/tmc/langchaingo/llms/vertexai/internal/aiplatformclient"
	"github.com/tmc/langchaingo/llms/vertexai/internal/genaiclient"
	"github.com/tmc/langchaingo/llms/vertexai/internal/vertexschema"
	"google.golang.org/api/option"
)

const (
	defaultMaxConns = 4
)

// VertexClient represents a Vertex AI API client.
type VertexClient struct {
	genAIClient  genaiclient.GenAIClient
	legacyClient aiplatformclient.PalmClient

	vertexschema.ConnectOptions
}

// New returns a new Vertex AI API client.
func New(ctx context.Context, opts vertexschema.ConnectOptions, copts ...option.ClientOption) (*VertexClient, error) {
	numConns := runtime.GOMAXPROCS(0)
	if numConns > defaultMaxConns {
		numConns = defaultMaxConns
	}
	o := []option.ClientOption{
		option.WithGRPCConnectionPool(numConns),
		option.WithEndpoint(opts.Endpoint),
	}
	o = append(o, copts...)

	client, err := genaiclient.New(ctx, opts.ProjectID, opts.Location, o...)
	if err != nil {
		return nil, err
	}

	legacyClient, err := aiplatformclient.New(ctx, opts.ProjectID, opts.Location, o...)
	if err != nil {
		return nil, err
	}

	return &VertexClient{
		genAIClient:    client,
		legacyClient:   legacyClient,
		ConnectOptions: opts,
	}, nil
}

// CreateCompletion creates a completion.
func (c *VertexClient) CreateCompletion(ctx context.Context, r *vertexschema.CompletionRequest) ([]*vertexschema.Completion, error) { //nolint:lll
	if strings.Contains(r.Model, "bison") || strings.Contains(r.Model, "gecko") {
		return c.legacyClient.CreateCompletion(ctx, r)
	}

	return c.genAIClient.CreateCompletion(ctx, r)
}

// EmbeddingRequest is a request to create an embedding.
type EmbeddingRequest struct {
	Input []string `json:"input"`
	Model string   `json:"model"`
}

// CreateEmbedding creates embeddings.
func (c *VertexClient) CreateEmbedding(ctx context.Context, r *EmbeddingRequest) ([][]float32, error) {
	model := r.Model

	params := map[string]interface{}{}
	responses, err := c.legacyClient.BatchPredict(ctx, model, r.Input, params)
	if err != nil {
		return nil, err
	}

	embeddings := [][]float32{}
	for _, res := range responses {
		value := res.GetStructValue().AsMap()
		embedding, ok := value["embeddings"].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("%w: %v", vertexschema.ErrMissingValue, "embeddings")
		}
		values, ok := embedding["values"].([]interface{})
		if !ok {
			return nil, fmt.Errorf("%w: %v", vertexschema.ErrMissingValue, "values")
		}
		floatValues := []float32{}
		for _, v := range values {
			val, ok := v.(float32)
			if !ok {
				valF64, ok := v.(float64)
				if !ok {
					return nil, fmt.Errorf("%w: %v is not a float64 or float32, it is a %T", vertexschema.ErrInvalidValue, "value", v)
				}
				val = float32(valF64)
			}
			floatValues = append(floatValues, val)
		}
		embeddings = append(embeddings, floatValues)
	}
	return embeddings, nil
}

func (c *VertexClient) CreateChat(ctx context.Context, model string, publisher string, r *vertexschema.ChatRequest) (*vertexschema.ChatResponse, error) { //nolint:lll
	if strings.Contains(model, "bison") || strings.Contains(model, "gecko") {
		return c.legacyClient.CreateChat(ctx, model, publisher, r)
	}

	return c.genAIClient.CreateChat(ctx, model, publisher, r)
}
