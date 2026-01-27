package anthropic

import (
	"github.com/vxcontrol/langchaingo/llms/anthropic/internal/anthropicclient"
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

	// Default cache strategy to apply to all requests unless overridden at call-time.
	defaultCacheStrategy *CacheStrategy
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

// WithDefaultCacheStrategy sets the default caching strategy for all requests made by this client.
// This strategy will be applied to all GenerateContent/Call invocations unless overridden
// by a call-level WithCacheStrategy option.
//
// This is safe and cost-effective when:
// - You have stable tools that don't change (CacheTools: true saves 90% on tools from 2nd request)
// - You use the same system prompt across conversations (CacheSystem: true)
// - You want automatic conversation history caching (CacheMessages: true for multi-turn)
//
// Example for AI agent with stable tools:
//
//	llm, err := anthropic.New(
//	    anthropic.WithDefaultCacheStrategy(anthropic.CacheStrategy{
//	        CacheTools: true,  // Tools cached once, reused forever
//	        CacheSystem: true, // System prompt cached once
//	    }),
//	)
//
// Call-level strategies override client-level on a per-field basis.
func WithDefaultCacheStrategy(strategy CacheStrategy) Option {
	return func(opts *options) {
		opts.defaultCacheStrategy = &strategy
	}
}
