package huggingface

import (
	"net/http"
)

const (
	tokenEnvVarName   = "HUGGINGFACEHUB_API_TOKEN" // Legacy environment variable
	hfTokenEnvVarName = "HF_TOKEN"                 // Current primary environment variable
	//nolint:gosec // This is not a hardcoded credential, it's an environment variable name
	hfTokenPathEnvVarName = "HF_TOKEN_PATH"  // Path to token file
	hfHomeEnvVarName      = "HF_HOME"        // HF home directory
	xdgCacheHomeEnvVar    = "XDG_CACHE_HOME" // XDG cache directory
	defaultTokenPath      = "token"          // Default token filename
	defaultModel          = "gpt2"
	defaultURL            = "https://api-inference.huggingface.co"
	routerURL             = "https://router.huggingface.co"
)

type options struct {
	token      string
	model      string
	url        string
	httpClient *http.Client
	provider   string // Inference provider (e.g., "hyperbolic", "nebius")
}

type Option func(*options)

// WithToken passes the HuggingFace API token to the client. If not set, the token
// is read from the HUGGINGFACEHUB_API_TOKEN environment variable.
func WithToken(token string) Option {
	return func(opts *options) {
		opts.token = token
	}
}

// WithModel passes the HuggingFace model to the client. If not set, then will be
// used default model.
func WithModel(model string) Option {
	return func(opts *options) {
		opts.model = model
	}
}

// WithURL passes the HuggingFace url to the client. If not set, then will be
// used default url.
func WithURL(url string) Option {
	return func(opts *options) {
		opts.url = url
	}
}

// WithHTTPClient passes a custom HTTP client to the HuggingFace client.
func WithHTTPClient(httpClient *http.Client) Option {
	return func(opts *options) {
		opts.httpClient = httpClient
	}
}

// WithInferenceProvider passes the inference provider to use with HuggingFace's router.
// When set, the client will use the router URL (https://router.huggingface.co/{provider}/v1/...)
// instead of the default inference API. Common providers include "hyperbolic", "nebius", etc.
func WithInferenceProvider(provider string) Option {
	return func(opts *options) {
		opts.provider = provider
	}
}
