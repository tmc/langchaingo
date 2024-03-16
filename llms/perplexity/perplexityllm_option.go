package perplexity

import (
	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/llms/perplexity/internal/perplexityclient"
)

const (
	tokenEnvVarName   = "PERPLEXITY_API_KEY"  //nolint:gosec
	modelEnvVarName   = "perplexity_MODEL"    //nolint:gosec
	baseURLEnvVarName = "perplexity_BASE_URL" //nolint:gosec
)

type options struct {
	token           string
	model           string
	baseURL         string
	httpClient      perplexityclient.Doer
	callbackHandler callbacks.Handler
}

type Option func(*options)

// WithToken passes the perplexity API token to the client. If not set, the token
// is read from the PERPLEXITY_API_KEY environment variable.
func WithToken(token string) Option {
	return func(opts *options) {
		opts.token = token
	}
}

// WithModel passes the perplexity model to the client. If not set, the model
// is read from the perplexity_MODEL environment variable.
func WithModel(model string) Option {
	return func(opts *options) {
		opts.model = model
	}
}

// WithBaseURL passes the perplexity base url to the client. If not set, the base url
// is read from the perplexity_BASE_URL environment variable. If still not set in ENV
// VAR perplexity_BASE_URL, then the default value is https://api.perplexity.com/v1 is used.
func WithBaseURL(baseURL string) Option {
	return func(opts *options) {
		opts.baseURL = baseURL
	}
}

// WithHTTPClient allows setting a custom HTTP client. If not set, the default value
// is http.DefaultClient.
func WithHTTPClient(client perplexityclient.Doer) Option {
	return func(opts *options) {
		opts.httpClient = client
	}
}

// WithCallback allows setting a custom Callback Handler.
func WithCallback(callbackHandler callbacks.Handler) Option {
	return func(opts *options) {
		opts.callbackHandler = callbackHandler
	}
}
