package azureaisearch

import (
	"errors"
	"net/http"
	"os"
	"strings"

	"github.com/vendasta/langchaingo/embeddings"
	"github.com/vendasta/langchaingo/vectorstores"
)

const (
	// EnvironmentVariableEndpoint environment variable to set azure ai search endpoint.
	EnvironmentVariableEndpoint string = "AZURE_AI_SEARCH_ENDPOINT"
	// EnvironmentVariableAPIKey environment variable to set azure ai api key.
	EnvironmentVariableAPIKey string = "AZURE_AI_SEARCH_API_KEY"
)

var (
	// ErrMissingEnvVariableAzureAISearchEndpoint environment variable to set azure ai search endpoint missing.
	ErrMissingEnvVariableAzureAISearchEndpoint = errors.New(
		"missing azureAISearchEndpoint",
	)
	// ErrMissingEmbedded embedder is missing, one should be set when instantiating the vectorstore.
	ErrMissingEmbedded = errors.New(
		"missing embedder",
	)
)

func (s *Store) getOptions(options ...vectorstores.Option) vectorstores.Options {
	opts := vectorstores.Options{}
	for _, opt := range options {
		opt(&opts)
	}
	return opts
}

// WithFilters can set the filter property in search document payload.
func WithFilters(filters any) vectorstores.Option {
	return func(o *vectorstores.Options) {
		o.Filters = filters
	}
}

// Option is a function type that can be used to modify the client.
type Option func(p *Store)

// WithEmbedder is an option for setting the embedder to use.
func WithEmbedder(e embeddings.Embedder) Option {
	return func(p *Store) {
		p.embedder = e
	}
}

// WithEmbedder is an option for setting the http client, the vectorstore uses the REST API,
// default http client is set but can be overridden by this option.
func WithHTTPClient(client *http.Client) Option {
	return func(s *Store) {
		s.client = client
	}
}

// WithAPIKey is an option for setting the azure AI search API Key.
func WithAPIKey(azureAISearchAPIKey string) Option {
	return func(s *Store) {
		s.azureAISearchAPIKey = azureAISearchAPIKey
	}
}

// WithEndpoint is an option for setting the azure AI search endpoint.
func WithEndpoint(endpoint string) Option {
	return func(s *Store) {
		s.azureAISearchEndpoint = strings.TrimSuffix(endpoint, "/")
	}
}

func applyClientOptions(s *Store, opts ...Option) error {
	for _, opt := range opts {
		opt(s)
	}

	if s.azureAISearchEndpoint == "" {
		s.azureAISearchEndpoint = strings.TrimSuffix(os.Getenv(EnvironmentVariableEndpoint), "/")
	}

	if s.azureAISearchEndpoint == "" {
		return ErrMissingEnvVariableAzureAISearchEndpoint
	}

	if s.embedder == nil {
		return ErrMissingEmbedded
	}

	if envVariableAPIKey := os.Getenv(EnvironmentVariableAPIKey); envVariableAPIKey != "" {
		s.azureAISearchAPIKey = envVariableAPIKey
	}

	return nil
}
