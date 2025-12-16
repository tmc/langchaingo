// package googleai implements a langchaingo provider for Google AI LLMs.
// See https://ai.google.dev/ for more details.
package googleai

import (
	"context"

	"github.com/google/generative-ai-go/genai"
	"github.com/vendasta/langchaingo/callbacks"
	"github.com/vendasta/langchaingo/llms"
)

// GoogleAI is a type that represents a Google AI API client.
type GoogleAI struct {
	CallbacksHandler callbacks.Handler
	client           *genai.Client
	opts             Options
	model            string // Track current model for reasoning detection
}

var (
	_ llms.Model          = &GoogleAI{}
	_ llms.ReasoningModel = &GoogleAI{}
)

// New creates a new GoogleAI client.
func New(ctx context.Context, opts ...Option) (*GoogleAI, error) {
	clientOptions := DefaultOptions()
	for _, opt := range opts {
		opt(&clientOptions)
	}
	clientOptions.EnsureAuthPresent()

	gi := &GoogleAI{
		opts:  clientOptions,
		model: clientOptions.DefaultModel, // Store the default model
	}

	client, err := genai.NewClient(ctx, clientOptions.ClientOptions...)
	if err != nil {
		return gi, err
	}

	gi.client = client
	return gi, nil
}

// Close closes the underlying genai client.
// This should be called when the GoogleAI instance is no longer needed
// to prevent memory leaks from the underlying gRPC connections.
func (g *GoogleAI) Close() error {
	if g.client != nil {
		return g.client.Close()
	}
	return nil
}

// SupportsReasoning implements the ReasoningModel interface.
// Returns false because the old SDK (github.com/google/generative-ai-go) does not
// support the ThinkingConfig API. For reasoning/thinking support, use googleaiv2.
func (g *GoogleAI) SupportsReasoning() bool {
	// The old SDK doesn't support ThinkingConfig API
	// Use googleaiv2 for models that require reasoning/thinking support
	return false
}
