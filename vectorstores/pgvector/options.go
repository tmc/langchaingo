package pgvector

import (
	"errors"
	"fmt"

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
		p.embeddingTableName = name
	}
}

// WithCollectionTableName is an option for specifying the collection table name.
func WithCollectionTableName(name string) Option {
	return func(p *Store) {
		p.collectionTableName = name
	}
}

// WithConn is an option for specifying the Postgres connection.
// From pgx doc: it is not safe for concurrent usage.Use a connection pool to manage access
// to multiple database connections from multiple goroutines.
func WithConn(conn PGXConn) Option {
	return func(p *Store) {
		p.conn = conn
	}
}

// WithCollectionMetadata is an option for specifying the collection metadata.
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

// WithHNSWIndex is an option for specifying the HNSW index parameters.
// See here for more details: https://github.com/pgvector/pgvector#hnsw
//
// m: he max number of connections per layer (16 by default)
// efConstruction: the size of the dynamic candidate list for constructing the graph (64 by default)
// distanceFunction: the distance function to use (l2 by default).
func WithHNSWIndex(m int, efConstruction int, distanceFunction string) Option {
	return func(p *Store) {
		p.hnswIndex = &HNSWIndex{
			m:                m,
			efConstruction:   efConstruction,
			distanceFunction: distanceFunction,
		}
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

	if o.conn == nil {
		return Store{}, fmt.Errorf("%w: missing postgres connection", ErrInvalidOptions)
	}

	if o.embedder == nil {
		return Store{}, fmt.Errorf("%w: missing embedder", ErrInvalidOptions)
	}

	return *o, nil
}
