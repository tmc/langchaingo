package azureaisearch

import (
	"errors"
	"net/http"
	"os"
	"strings"

	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/vectorstores"
)

const (
	EnvironmentVariableEndpoint string = "COGNITIVE_SEARCH_ENDPOINT"
	EnvironmentVariableAPIKey   string = "COGNITIVE_SEARCH_API_KEY"
)

var (
	ErrMissingEnvVariableCognitiveSearchEndpoint = errors.New(
		"missing cognitiveSearchEndpoint",
	)
	ErrMissingEmbedded = errors.New(
		"missing embedder",
	)
)

func (s Store) getOptions(options ...vectorstores.Option) vectorstores.Options {
	opts := vectorstores.Options{}
	for _, opt := range options {
		opt(&opts)
	}
	return opts
}

func WithFilters(filters any) vectorstores.Option {
	return func(o *vectorstores.Options) {
		o.Filters = filters
	}
}

type Option func(p *Store)

func WithEmbedder(e embeddings.Embedder) Option {
	return func(p *Store) {
		p.embedder = e
	}
}

func WithHTTPClient(client *http.Client) Option {
	return func(s *Store) {
		s.client = client
	}
}

func WithAPIKey(cognitiveSearchAPIKey string) Option {
	return func(s *Store) {
		s.cognitiveSearchAPIKey = cognitiveSearchAPIKey
	}
}

func applyClientOptions(s *Store, opts ...Option) error {
	for _, opt := range opts {
		opt(s)
	}

	if s.cognitiveSearchEndpoint == "" {
		s.cognitiveSearchEndpoint = strings.TrimSuffix(os.Getenv(EnvironmentVariableEndpoint), "/")
	}

	if s.cognitiveSearchEndpoint == "" {
		return ErrMissingEnvVariableCognitiveSearchEndpoint
	}

	if s.embedder == nil {
		return ErrMissingEmbedded
	}

	if envVariableAPIKey := os.Getenv(EnvironmentVariableAPIKey); envVariableAPIKey != "" {
		s.cognitiveSearchAPIKey = envVariableAPIKey
	}

	return nil
}
