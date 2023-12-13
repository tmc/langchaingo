package pgvector

import (
	"errors"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/tmc/langchaingo/embeddings"
)

const (
	DefaultCollectionName           = "langchain"
	DefaultPreDeleteCollection      = false
	DefaultEmbeddingStoreTableName  = "langchain_pg_embedding"
	DefaultCollectionStoreTableName = "langchain_pg_collection"
)

// ErrInvalidOptions is returned when the options given are invalid.
var ErrInvalidOptions = errors.New("invalid options")

// Option is a function type that can be used to modify the client.
type Option func(p *Store)

// WithEmbedder is an option for setting the embedder to use. Must be set.
func WithEmbedder(e embeddings.Embedder) Option {
	return func(p *Store) {
		p.embedder = e
	}
}

// WithConnectionURL is an option for specifying the Postgres connection URL. Must be set.
func WithConnectionURL(connectionURL string) Option {
	return func(p *Store) {
		p.postgresConnectionURL = connectionURL
	}
}

// WithPreDeleteCollection is an option for setting if the collection should be deleted before creating.
func WithPreDeleteCollection(preDelete bool) Option {
	return func(p *Store) {
		p.preDeleteCollection = preDelete
	}
}

// WithCollectionName is an option for specifying the collection name.
func WithCollectionName(name string) Option {
	return func(p *Store) {
		p.collectionName = name
	}
}

// WithEmbeddingTableName is an option for specifying the embedding table name.
func WithEmbeddingTableName(name string) Option {
	return func(p *Store) {
		p.embeddingTableName = pgx.Identifier{name}.Sanitize()
	}
}

// WithCollectionTableName is an option for specifying the collection table name.
func WithCollectionTableName(name string) Option {
	return func(p *Store) {
		p.collectionTableName = pgx.Identifier{name}.Sanitize()
	}
}

func applyClientOptions(opts ...Option) (Store, error) {
	o := &Store{
		collectionName:      DefaultCollectionName,
		preDeleteCollection: DefaultPreDeleteCollection,
		embeddingTableName:  DefaultEmbeddingStoreTableName,
		collectionTableName: DefaultCollectionStoreTableName,
	}

	for _, opt := range opts {
		opt(o)
	}

	if o.postgresConnectionURL == "" {
		o.postgresConnectionURL = os.Getenv("PGVECTOR_CONNECTION_STRING")
	}

	if o.postgresConnectionURL == "" {
		return Store{}, fmt.Errorf("%w: missing postgresConnectionURL", ErrInvalidOptions)
	}

	if o.embedder == nil {
		return Store{}, fmt.Errorf("%w: missing embedder", ErrInvalidOptions)
	}

	return *o, nil
}
