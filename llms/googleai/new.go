// package googleai implements a langchaingo provider for Google AI LLMs.
// See https://ai.google.dev/ for more details.
package googleai

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/vendasta/langchaingo/callbacks"
	"github.com/vendasta/langchaingo/llms"
	"google.golang.org/genai"
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

	// Build ClientConfig for the new SDK
	clientConfig := &genai.ClientConfig{}

	// Use API key from Options if available
	if clientOptions.APIKey != "" {
		clientConfig.APIKey = clientOptions.APIKey
		clientConfig.Backend = genai.BackendGeminiAPI
	} else if apiKey := os.Getenv("GOOGLE_API_KEY"); apiKey != "" {
		// Fall back to environment variable
		clientConfig.APIKey = apiKey
		clientConfig.Backend = genai.BackendGeminiAPI
	} else {
		// For now, require API key for Google AI (Vertex AI handled separately)
		return gi, fmt.Errorf("API key required for Google AI client")
	}

	client, err := genai.NewClient(ctx, clientConfig)
	if err != nil {
		return gi, err
	}

	gi.client = client
	return gi, nil
}

// Close closes the underlying genai client.
// The new SDK's Client doesn't expose a Close method as it uses HTTP clients
// that are managed internally and don't require explicit cleanup.
// This method is provided for API compatibility and to match the interface
// expected by callers who may be used to closing clients.
func (g *GoogleAI) Close() error {
	return nil
}

// SupportsReasoning implements the ReasoningModel interface.
// Returns true if the current model supports reasoning/thinking tokens.
func (g *GoogleAI) SupportsReasoning() bool {
	// Check the current model (may have been overridden by WithModel option)
	model := g.model
	if model == "" {
		model = g.opts.DefaultModel
	}

	// Gemini 2.0 models support reasoning/thinking capabilities
	if strings.Contains(model, "gemini-2.0") {
		return true
	}

	// Future Gemini 3+ models expected to support reasoning
	if strings.Contains(model, "gemini-3") || strings.Contains(model, "gemini-4") {
		return true
	}

	// Gemini Experimental models may have reasoning capabilities
	if strings.Contains(model, "gemini-exp") && strings.Contains(model, "thinking") {
		return true
	}

	return false
}
