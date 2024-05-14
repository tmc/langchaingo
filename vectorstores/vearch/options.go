package vearch

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/tmc/langchaingo/embeddings"
)

const (
	DefaultDbName    = "ts_db"
	DefaultSpaceName = "ts_space"
)

// ErrInvalidOptions is returned when the options given are invalid.
var ErrInvalidOptions = errors.New("invalid options")

// Option is a function that configures an Options.
type Option func(store *Store)

// WithEmbedder is an option for setting the embedder to use. Must be set.
func WithEmbedder(embedder embeddings.Embedder) Option {
	return func(store *Store) {
		store.embedder = embedder
	}
}

// WithDbName returns an Option for setting the database name. Required.
func WithDbName(name string) Option {
	return func(store *Store) {
		store.DbName = name
	}
}

// WithSpaceName returns an Option for setting the space name. Required.
func WithSpaceName(name string) Option {
	return func(store *Store) {
		store.SpaceName = name
	}
}

// WithURL returns an Option for setting the Vearch cluster URL. Required.
func WithURL(clusterUrl url.URL) Option {
	return func(store *Store) {
		store.ClusterUrl = clusterUrl
	}
}

func applyClientOptions(opts ...Option) (Store, error) {
	store := Store{}

	for _, opt := range opts {
		opt(&store)
	}

	if store.DbName == "" {
		return Store{}, fmt.Errorf("%w: missing database name", ErrInvalidOptions)
	}

	if store.SpaceName == "" {
		return Store{}, fmt.Errorf("%w: missing space name", ErrInvalidOptions)
	}

	if store.ClusterUrl == (url.URL{}) {
		return Store{}, fmt.Errorf("%w: missing cluster URL", ErrInvalidOptions)
	}

	if store.embedder == nil {
		return Store{}, fmt.Errorf("%w: missing embedder", ErrInvalidOptions)
	}

	return store, nil
}
