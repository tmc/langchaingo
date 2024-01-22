package vertex

import (
	"context"
	"log"

	"cloud.google.com/go/vertexai/genai"
	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/llms"
)

// Vertex is a type that represents a Vertex AI API client.
//
// TODO: This isn't in common code; may need PaLM client for embeddings, etc.
// Note the deltas: type of topk, candidate count.
type Vertex struct {
	CallbacksHandler callbacks.Handler
	client           *genai.Client
	opts             options
}

var _ llms.Model = &Vertex{}

// NewVertex creates a new Vertex struct.
func NewVertex(ctx context.Context, opts ...Option) (*Vertex, error) {
	clientOptions := defaultOptions()
	for _, opt := range opts {
		opt(&clientOptions)
	}

	v := &Vertex{
		opts: clientOptions,
	}

	client, err := genai.NewClient(ctx, clientOptions.cloudProject, clientOptions.cloudLocation)
	if err != nil {
		log.Fatal(err)
	}

	v.client = client
	return v, nil
}
