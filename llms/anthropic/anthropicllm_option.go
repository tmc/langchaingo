package anthropic

import (
	"github.com/tmc/langchaingo/llms/anthropic/internal/anthropicclient"
)

const (
	tokenEnvVarName = "ANTHROPIC_API_KEY" //nolint:gosec
)

// MaxTokensAnthropicSonnet35 is the header value for specifying the maximum number of tokens
// when using the Anthropic Sonnet 3.5 model.
const MaxTokensAnthropicSonnet35 = "max-tokens-3-5-sonnet-2024-07-15" //nolint:gosec // This is not a sensitive value.

type options struct {
	token      string
	model      string
	baseURL    string
	httpClient anthropicclient.Doer

	useLegacyTextCompletionsAPI bool

	// If supplied, the 'anthropic-beta' header will be added to the request with the given value.
	anthropicBetaHeader string

	cacheTools         bool
	cacheSystemMessage bool
	cacheChat          bool
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

// WithLegacyTextCompletionsAPI enables the use of the legacy text completions API.
func WithLegacyTextCompletionsAPI() Option {
	return func(opts *options) {
		opts.useLegacyTextCompletionsAPI = true
	}
}

// WithAnthropicBetaHeader adds the Anthropic Beta header to support extended options.
func WithAnthropicBetaHeader(value string) Option {
	return func(opts *options) {
		opts.anthropicBetaHeader = value
	}
}

// WithCacheTools enables caching of tool definitions. See https://docs.anthropic.com/en/docs/build-with-claude/prompt-caching
func WithCacheTools() Option {
	return func(opts *options) {
		opts.cacheTools = true
	}
}

// WithCacheSystemMessage enables caching of the system message. See https://docs.anthropic.com/en/docs/build-with-claude/prompt-caching
func WithCacheSystemMessage() Option {
	return func(opts *options) {
		opts.cacheSystemMessage = true
	}
}

// WithCacheChat enables caching of chat messages. See https://docs.anthropic.com/en/docs/build-with-claude/prompt-caching
func WithCacheChat() Option {
	return func(opts *options) {
		opts.cacheChat = true
	}
}
