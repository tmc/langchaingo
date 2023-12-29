package opensearch

import (
	"errors"

	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/vectorstores"
)

var (
	ErrMissingEnvVariableOpensearchEndpoint = errors.New(
		"missing opensearchEndpoint",
	)
	ErrMissingEmbedded = errors.New(
		"missing embedder",
	)
	ErrMissingOpensearchClient = errors.New(
		"missing opensearch client",
	)
)

func (s Store) getOptions(options ...vectorstores.Option) vectorstores.Options {
	opts := vectorstores.Options{}
	for _, opt := range options {
		opt(&opts)
	}
	return opts
}

// Option is a function type that can be used to modify the client.
type Option func(p *Store)

// WithEmbedder returns an Option for setting the embedder that could be used when
// adding documents or doing similarity search (instead the embedder from the Store context)
// this is useful when we are using multiple LLMs with single vectorstore.
func WithEmbedder(e embeddings.Embedder) Option {
	return func(p *Store) {
		p.embedder = e
	}
}

func applyClientOptions(s *Store, opts ...Option) error {
	for _, opt := range opts {
		opt(s)
	}

	if s.embedder == nil {
		return ErrMissingEmbedded
	}

	if s.client == nil {
		return ErrMissingEmbedded
	}

	return nil
}
