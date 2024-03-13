package anthropic

import (
	"github.com/tmc/langchaingo/llms/anthropic/internal/anthropicclient"
)

const (
	tokenEnvVarName = "ANTHROPIC_API_KEY" //nolint:gosec
)

type options struct {
	token      string
	model      string
	baseURL    string
	httpClient anthropicclient.Doer
}

type Option func(*options)

// WithToken passes the Anthropic API token to the client. If not set, the token
// is read from the ANTHROPIC_API_KEY environment variable.
func WithToken(token string) Option {
	return func(opts *options) {
		opts.token = token
	}
}

// WithModel passes the Anthropic model to the client.
func WithModel(model string) Option {
	return func(opts *options) {
		opts.model = model
	}
}

// WithBaseUrl passes the Anthropic base URL to the client.
// If not set, the default base URL is used.
func WithBaseURL(baseURL string) Option {
	return func(opts *options) {
		opts.baseURL = baseURL
	}
}

// WithHTTPClient allows setting a custom HTTP client. If not set, the default value
// is http.DefaultClient.
func WithHTTPClient(client anthropicclient.Doer) Option {
	return func(opts *options) {
		opts.httpClient = client
	}
}
