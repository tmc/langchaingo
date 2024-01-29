package pgvector

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/tmc/langchaingo/embeddings"
)

const (
	DefaultCollectionName           = "langchain"
	DefaultPreDeleteCollection      = false
	DefaultEmbeddingStoreTableName  = "langchain_pg_embedding"
	DefaultCollectionStoreTableName = "langchain_pg_collection"
	DefaultVectorDimensions         = 1536
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
		p.embeddingTableName = tableName(name)
	}
}

// WithCollectionTableName is an option for specifying the collection table name.
func WithCollectionTableName(name string) Option {
	return func(p *Store) {
		p.collectionTableName = tableName(name)
	}
}

// WithConn is an option for specifying the Postgres connection.
// From pgx doc: it is not safe for concurrent usage.Use a connection pool to manage access
// to multiple database connections from multiple goroutines.
func WithConn(conn *pgx.Conn) Option {
	return func(p *Store) {
		p.conn = conn
	}
}

func WithCollectionMetadata(metadata map[string]any) Option {
	return func(p *Store) {
		p.collectionMetadata = metadata
	}
}

// WithVectorDimensions is an option for specifying the vector size.
func WithVectorDimensions(size int) Option {
	return func(p *Store) {
		p.vectorDimensions = size
	}
}

func applyClientOptions(opts ...Option) (Store, error) {
	o := &Store{
		collectionName:      DefaultCollectionName,
		preDeleteCollection: DefaultPreDeleteCollection,
		embeddingTableName:  DefaultEmbeddingStoreTableName,
		collectionTableName: DefaultCollectionStoreTableName,
		vectorDimensions:    DefaultVectorDimensions,
	}

	for _, opt := range opts {
		opt(o)
	}

	if o.postgresConnectionURL == "" {
		o.postgresConnectionURL = os.Getenv("PGVECTOR_CONNECTION_STRING")
	}

	if o.postgresConnectionURL == "" && o.conn == nil {
		return Store{}, fmt.Errorf("%w: missing postgresConnectionURL", ErrInvalidOptions)
	}

	if o.embedder == nil {
		return Store{}, fmt.Errorf("%w: missing embedder", ErrInvalidOptions)
	}

	return *o, nil
}

// tableName returns the table name with the schema sanitized.
func tableName(name string) string {
	nameParts := strings.Split(name, ".")

	return pgx.Identifier(nameParts).Sanitize()
}
