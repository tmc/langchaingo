package gigachat

import (
	"github.com/tmc/langchaingo/llms/gigachat/internal/gigachatclient"
)

const (
	clientIdEnvVarName     = "GIGACHAT_CLIENT_ID"     //nolint:gosec
	clientSecretEnvVarName = "GIGACHAT_CLIENT_SECRET" //nolint:gosec
)

type options struct {
	clientId     string
	clientSecret string
	scope        string
	model        string
	baseURL      string
	httpClient   gigachatclient.Doer
}

type Option func(*options)

// WithClientIdAndSecret passes the Gigachat API creds to the client. If not set, the creds
// are read from the GIGACHAT_CLIENT_ID and GIGACHAT_CLIENT_SECRET environment variables.
func WithClientIdAndSecret(clientId, clientSecret string) Option {
	return func(opts *options) {
		opts.clientId = clientId
		opts.clientSecret = clientSecret
	}
}

// WithScope passes the scope to the client.
func WithScope(scope string) Option {
	return func(opts *options) {
		opts.scope = scope
	}
}

// WithModel passes the Gigachat model to the client.
func WithModel(model string) Option {
	return func(opts *options) {
		opts.model = model
	}
}

// WithBaseUrl passes the Gigachat base URL to the client.
// If not set, the default base URL is used.
func WithBaseURL(baseURL string) Option {
	return func(opts *options) {
		opts.baseURL = baseURL
	}
}

// WithHTTPClient allows setting a custom HTTP client. If not set, the default value
// is http.DefaultClient.
func WithHTTPClient(client gigachatclient.Doer) Option {
	return func(opts *options) {
		opts.httpClient = client
	}
}
