package mariadb

import (
	"errors"
	"fmt"

	"github.com/tmc/langchaingo/embeddings"
)

const (
	DefaultCollectionName           = "langchain"
	DefaultPreDeleteCollection      = false
	DefaultEmbeddingStoreTableName  = "langchain_embedding"
	DefaultCollectionStoreTableName = "langchain_collection"
	DefaultVectorSize               = 1536
)

// ErrInvalidOptions is returned when the options given are invalid.
var ErrInvalidOptions = errors.New("invalid options")

// Option is a function type that can be used to modify the store.
type Option func(p *Store)

// WithDSN is an option for specifying the MariaDB DSN.
func WithDSN(dsn string) Option {
	return func(p *Store) {
		p.dsn = dsn
	}
}

// WithEmbedder is an option for setting the embedder to use. Must be set.
func WithEmbedder(e embeddings.Embedder) Option {
	return func(p *Store) {
		p.embedder = e
	}
}

// WithHNSWIndex is an option for configuring the HNSW index.
// Allows overriding the default parameters.
// See here for more details: https://mariadb.com/kb/en/create-table-with-vectors/
func WithHNSWIndex(m int, distance DistanceFunction) Option {
	return func(p *Store) {
		p.hnswIndex = &HNSWIndex{
			M:            m,
			DistanceFunc: distance,
		}
	}
}

// WithVectorDimensions is an option for specifying the vector size.
func WithVectorDimensions(size int) Option {
	return func(p *Store) {
		p.vectorDimensions = size
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

// WithPreDeleteCollection is an option for setting if the collection should be deleted before creating.
func WithPreDeleteCollection(preDelete bool) Option {
	return func(p *Store) {
		p.preDeleteCollection = preDelete
	}
}

// WithCollectionMetadata is an option for specifying the collection metadata.
func WithCollectionMetadata(metadata map[string]any) Option {
	return func(p *Store) {
		p.collectionMetadata = metadata
	}
}

func applyClientOptions(opts ...Option) (*Store, error) {
	store := &Store{
		collectionName:      DefaultCollectionName,
		embeddingTableName:  DefaultEmbeddingStoreTableName,
		collectionTableName: DefaultCollectionStoreTableName,
		vectorDimensions:    DefaultVectorSize,
	}

	for _, opt := range opts {
		opt(store)
	}

	if store.dsn == "" {
		return nil, fmt.Errorf("%w: missing dsn", ErrInvalidOptions)
	}

	if store.hnswIndex == nil {
		store.hnswIndex = &HNSWIndex{
			M:            8,
			DistanceFunc: Cosine,
		}
	}

	if store.hnswIndex.M <= 0 {
		return nil, fmt.Errorf("%w: HNSW M must be > 0", ErrInvalidOptions)
	}

	if store.embedder == nil {
		return nil, fmt.Errorf("%w: missing embedder", ErrInvalidOptions)
	}

	if store.vectorDimensions <= 0 {
		return nil, fmt.Errorf("%w: vector dimensions must be > 0", ErrInvalidOptions)
	}

	return store, nil
}
