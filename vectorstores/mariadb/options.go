package mariadb

import (
	"errors"
	"fmt"

	"github.com/vendasta/langchaingo/embeddings"
)

const (
	DefaultDatabaseName             = "langchain"
	DefaultPreDeleteDatabase        = false
	DefaultEmbeddingStoreTableName  = "langchain_mariadb_embedding"
	DefaultCollectionStoreTableName = "langchain_mariadb_collection"
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

// WithPreDeleteDatabase is an option for setting if the database should be deleted before creating.
func WithPreDeleteDatabase(preDelete bool) Option {
	return func(p *Store) {
		p.preDeleteDatabase = preDelete
	}
}

// WithDatabaseName is an option for specifying the database name.
func WithDatabaseName(name string) Option {
	return func(p *Store) {
		p.databaseName = name
	}
}

// WithEmbeddingTableName is an option for specifying the embedding table name.
func WithEmbeddingTableName(name string) Option {
	return func(p *Store) {
		p.embeddingTableName = name
	}
}

// WithCollectionTableName is an option for specifying the collection table name.
func WithCollectionTableName(name string) Option {
	return func(p *Store) {
		p.collectionTableName = name
	}
}

// WithConnectionURL is an option for specifying the MariaDB connection URL. Either this
// or WithConn must be used.
func WithConnectionURL(connectionURL string) Option {
	return func(p *Store) {
		p.connURL = connectionURL
	}
}

// WithDB is an option for specifying the MariaDB connection.
func WithDB(db DB) Option {
	return func(p *Store) {
		p.db = db
	}
}

// WithDatabaseMetadata is an option for specifying the database metadata.
func WithDatabaseMetadata(metadata map[string]any) Option {
	return func(p *Store) {
		p.databaseMetadata = metadata
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
		databaseName:        DefaultDatabaseName,
		preDeleteDatabase:   DefaultPreDeleteDatabase,
		embeddingTableName:  DefaultEmbeddingStoreTableName,
		collectionTableName: DefaultCollectionStoreTableName,
	}

	for _, opt := range opts {
		opt(o)
	}

	if o.db == nil && o.connURL == "" {
		return Store{}, fmt.Errorf("%w: missing mariadb connection", ErrInvalidOptions)
	}

	if o.embedder == nil {
		return Store{}, fmt.Errorf("%w: missing embedder", ErrInvalidOptions)
	}

	return *o, nil
}
