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
