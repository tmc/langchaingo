package azureaisearch

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/vectorstores"
)

const EnvironmentVariable_Endpoint string = "COGNITIVE_SEARCH_ENDPOINT"
const EnvironmentVariable_APIKey string = "COGNITIVE_SEARCH_API_KEY"

func (s Store) getOptions(options ...vectorstores.Option) vectorstores.Options {
	opts := vectorstores.Options{}
	for _, opt := range options {
		opt(&opts)
	}
	return opts
}

func WithFilters(filters SimilaritySearchFilters) vectorstores.Option {
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
		s.cognitiveSearchEndpoint = strings.TrimSuffix(os.Getenv(EnvironmentVariable_Endpoint), "/")
	}

	if s.cognitiveSearchEndpoint == "" {
		return fmt.Errorf("missing cognitiveSearchEndpoint")
	}

	if s.embedder == nil {
		return fmt.Errorf("missing embedder")
	}

	if envVariableAPIKey := os.Getenv(EnvironmentVariable_APIKey); envVariableAPIKey != "" {
		s.cognitiveSearchAPIKey = envVariableAPIKey
	}

	return nil
}
