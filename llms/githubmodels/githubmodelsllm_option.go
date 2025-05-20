package githubmodels

import (
	"net/http"

	"github.com/tmc/langchaingo/callbacks"
)

const (
	// tokenEnvVarName is the environment variable name for the GitHub token.
	tokenEnvVarName = "GITHUB_TOKEN"
	// defaultModel is the default GitHub model to use.
	defaultModel = "openai/gpt-4.1"
)

// Available GitHub Models (as of May 2025):
// - openai/gpt-4.1 - OpenAI's GPT-4.1 model
// - anthropic/claude-3-sonnet - Anthropic's Claude 3 Sonnet model
// - anthropic/claude-3-haiku - Anthropic's Claude 3 Haiku model
// - mistral/mistral-large - Mistral's Large model
// - mistral/mistral-small - Mistral's Small model
// 
// Note: Available models may change over time. Check GitHub documentation for the latest information.

// options contains options for the GitHub Models client.
type options struct {
	token           string
	model           string
	httpClient      *http.Client
	callbacksHandler callbacks.Handler
}

// Option is a function that configures the GitHub Models client.
type Option func(*options)

// WithToken sets the GitHub token for the client.
// If not set, then will be used the GITHUB_TOKEN environment variable.
func WithToken(token string) Option {
	return func(o *options) {
		o.token = token
	}
}

// WithModel sets the model to use.
// Default is "openai/gpt-4.1".
func WithModel(model string) Option {
	return func(o *options) {
		o.model = model
	}
}

// WithHTTPClient sets the HTTP client to use.
// Default is http.DefaultClient.
func WithHTTPClient(client *http.Client) Option {
	return func(o *options) {
		o.httpClient = client
	}
}

// WithCallbacksHandler sets the callbacks handler.
func WithCallbacksHandler(handler callbacks.Handler) Option {
	return func(o *options) {
		o.callbacksHandler = handler
	}
}
