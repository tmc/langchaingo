// package googleai implements a langchaingo provider for Google AI LLMs.
// See https://ai.google.dev/ for more details.
package googleai

import (
	"context"

	"github.com/google/generative-ai-go/genai"
	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/llms"
	"google.golang.org/api/option"
)

// GoogleAI is a type that represents a Google AI API client.
type GoogleAI struct {
	CallbacksHandler callbacks.Handler
	client           *genai.Client
	opts             options
}

var _ llms.Model = &GoogleAI{}

// NewGoogleAI creates a new GoogleAI struct.
func NewGoogleAI(ctx context.Context, opts ...Option) (*GoogleAI, error) {
	clientOptions := defaultOptions()
	for _, opt := range opts {
		opt(&clientOptions)
	}

	gi := &GoogleAI{
		opts: clientOptions,
	}

	client, err := genai.NewClient(ctx, option.WithAPIKey(clientOptions.apiKey))
	if err != nil {
		return gi, err
	}

	gi.client = client
	return gi, nil
}
