// package vertex implements a langchaingo provider for Google Vertex AI LLMs,
// including the new Gemini models.
// See https://cloud.google.com/vertex-ai for more details.
package vertex

import (
	"context"

	"cloud.google.com/go/vertexai/genai"
	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/googleai/internal/palmclient"
)

// Vertex is a type that represents a Vertex AI API client.
//
// Right now, the Vertex Gemini SDK doesn't support embeddings; therefore,
// for embeddings we also hold a palmclient.
type Vertex struct {
	CallbacksHandler callbacks.Handler
	client           *genai.Client
	opts             options
	palmClient       *palmclient.PaLMClient
}

var _ llms.Model = &Vertex{}

// NewVertex creates a new Vertex client.
func NewVertex(ctx context.Context, opts ...Option) (*Vertex, error) {
	clientOptions := defaultOptions()
	for _, opt := range opts {
		opt(&clientOptions)
	}

	client, err := genai.NewClient(ctx, clientOptions.cloudProject, clientOptions.cloudLocation)
	if err != nil {
		return nil, err
	}

	palmClient, err := palmclient.New(clientOptions.cloudProject) //nolint:contextcheck
	if err != nil {
		return nil, err
	}

	v := &Vertex{
		opts:       clientOptions,
		client:     client,
		palmClient: palmClient,
	}
	return v, nil
}
