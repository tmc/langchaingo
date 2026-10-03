// package googleai implements a langchaingo provider for Google AI LLMs.
// See https://ai.google.dev/ for more details.
package googleai

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/llms"
	"google.golang.org/api/option"
)

type apiKeyTransport struct {
	base http.RoundTripper
	key  string
}

func (t apiKeyTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	clone := req.Clone(req.Context())
	q := clone.URL.Query()
	q.Set("key", t.key)
	clone.URL.RawQuery = q.Encode()
	return t.base.RoundTrip(clone)
}

func withAPIKey(client *http.Client, key string) *http.Client {
	clone := *client
	base := clone.Transport
	if base == nil {
		base = http.DefaultTransport
	}
	clone.Transport = apiKeyTransport{base: base, key: key}
	return &clone
}

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
	if clientOptions.auth == authAPIKey && clientOptions.apiKey != "" && clientOptions.httpClient != nil {
		clientOptions.ClientOptions = append(clientOptions.ClientOptions,
			option.WithHTTPClient(withAPIKey(clientOptions.httpClient, clientOptions.apiKey)))
	}

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
// Returns true if the current model supports reasoning/thinking tokens.
func (g *GoogleAI) SupportsReasoning() bool {
	// Check the current model (may have been overridden by WithModel option)
	model := g.model
	if model == "" {
		model = g.opts.DefaultModel
	}

	// Gemini 2.0 and 2.5 models support reasoning/thinking capabilities
	if strings.Contains(model, "gemini-2.0") || strings.Contains(model, "gemini-2.5") {
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
