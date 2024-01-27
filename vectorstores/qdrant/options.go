package qdrant

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/tmc/langchaingo/embeddings"
)

const (
	defaultContentKey = "content"
)

// ErrInvalidOptions is returned when the options given are invalid.
var ErrInvalidOptions = errors.New("invalid options")

// Option is a function that configures an Options.
type Option func(p *Store)

// WithCollectionName returns an Option for setting the collection name. Required.
func WithCollectionName(name string) Option {
	return func(p *Store) {
		p.collectionName = name
	}
}

// WithURL returns an Option for setting the Qdrant instance URL.
// Example: 'http://localhost:63333'. Required.
func WithURL(qdrantURL url.URL) Option {
	return func(p *Store) {
		p.qdrantURL = qdrantURL
	}
}

// WithEmbedder returns an Option for setting the embedder to be used when
// adding documents or doing similarity search. Required.
func WithEmbedder(embedder embeddings.Embedder) Option {
	return func(p *Store) {
		p.embedder = embedder
	}
}

// WithAPIKey returns an Option for setting the API key to authenticate the connection. Optional.
func WithAPIKey(apiKey string) Option {
	return func(p *Store) {
		p.apiKey = apiKey
	}
}

// WithContent returns an Option for setting field name of the document content
// in the Qdrant payload. Optional. Defaults to "content".
func WithContentKey(contentKey string) Option {
	return func(p *Store) {
		p.contentKey = contentKey
	}
}

func applyClientOptions(opts ...Option) (Store, error) {
	o := &Store{
		contentKey: defaultContentKey,
	}

	for _, opt := range opts {
		opt(o)
	}

	if o.collectionName == "" {
		return Store{}, fmt.Errorf("%w: missing collection name", ErrInvalidOptions)
	}

	if o.qdrantURL == (url.URL{}) {
		return Store{}, fmt.Errorf("%w: missing Qdrant URL", ErrInvalidOptions)
	}

	if o.embedder == nil {
		return Store{}, fmt.Errorf("%w: missing embedder", ErrInvalidOptions)
	}

	return *o, nil
}
