// package vertex implements a langchaingo provider for Google Vertex AI LLMs,
// including the new Gemini models.
// See https://cloud.google.com/vertex-ai for more details.
package vertex

import (
	"context"

	"cloud.google.com/go/vertexai/genai"
	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/googleai"
)

// Vertex is a type that represents a Vertex AI API client.
//
// For embeddings, we use the aiplatform API directly since the Vertex Gemini SDK
// doesn't have native embedding support yet. This provides access to Gemini
// embedding models like text-embedding-004 and gemini-embedding-001.
type Vertex struct {
	CallbacksHandler callbacks.Handler
	client           *genai.Client
	opts             googleai.Options
}

var _ llms.Model = &Vertex{}

// New creates a new Vertex client.
func New(ctx context.Context, opts ...googleai.Option) (*Vertex, error) {
	clientOptions := googleai.DefaultOptions()
	for _, opt := range opts {
		opt(&clientOptions)
	}

	client, err := genai.NewClient(
		ctx,
		clientOptions.CloudProject,
		clientOptions.CloudLocation,
		clientOptions.ClientOptions...)
	if err != nil {
		return nil, err
	}

	v := &Vertex{
		opts:   clientOptions,
		client: client,
	}
	return v, nil
}

// Close closes the underlying genai client.
// This should be called when the Vertex instance is no longer needed
// to prevent memory leaks from the underlying gRPC connections.
func (v *Vertex) Close() error {
	if v.client != nil {
		return v.client.Close()
	}
	return nil
}
