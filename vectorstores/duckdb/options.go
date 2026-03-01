package duckdb

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/tmc/langchaingo/embeddings"
)

const (
	DefaultCollectionName           = "langchain"
	DefaultPreDeleteCollection      = false
	DefaultEmbeddingStoreTableName  = "langchain_duckdb_embedding"
	DefaultCollectionStoreTableName = "langchain_duckdb_collection"
	DefaultDistanceMetric           = "cosine"
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

// WithConnectionURL is an option for specifying the DuckDB connection URL.
// Use an empty string for an in-memory database. Either this or WithDB must be used.
func WithConnectionURL(connectionURL string) Option {
	return func(p *Store) {
		p.connURL = connectionURL
		p.connURLSet = true
	}
}

// WithDB is an option for specifying a pre-existing DuckDB connection.
// When using this option, the vss extension must already be loaded by the caller.
func WithDB(db *sql.DB) Option {
	return func(p *Store) {
		p.db = db
	}
}

// WithCollectionName is an option for specifying the collection name.
func WithCollectionName(name string) Option {
	return func(p *Store) {
		p.collectionName = name
	}
}

// WithCollectionTableName is an option for specifying the collection table name.
func WithCollectionTableName(name string) Option {
	return func(p *Store) {
		p.collectionTableName = name
	}
}

// WithEmbeddingTableName is an option for specifying the embedding table name.
func WithEmbeddingTableName(name string) Option {
	return func(p *Store) {
		p.embeddingTableName = name
	}
}

// WithCollectionMetadata is an option for specifying the collection metadata.
func WithCollectionMetadata(metadata map[string]any) Option {
	return func(p *Store) {
		p.collectionMetadata = metadata
	}
}

// WithPreDeleteCollection is an option for setting if the collection should be deleted before creating.
func WithPreDeleteCollection(preDelete bool) Option {
	return func(p *Store) {
		p.preDeleteCollection = preDelete
	}
}

// WithVectorDimensions is an option for specifying the vector size. Must be set.
// DuckDB requires a fixed dimension for FLOAT[N] array columns.
func WithVectorDimensions(size int) Option {
	return func(p *Store) {
		p.vectorDimensions = size
	}
}

// WithDistanceMetric is an option for specifying the distance metric used
// for the HNSW index and similarity search queries.
// Supported values: "cosine" (default), "l2sq", "ip".
func WithDistanceMetric(metric string) Option {
	return func(p *Store) {
		p.distanceMetric = metric
	}
}

func applyClientOptions(opts ...Option) (Store, error) {
	o := &Store{
		collectionName:      DefaultCollectionName,
		preDeleteCollection: DefaultPreDeleteCollection,
		embeddingTableName:  DefaultEmbeddingStoreTableName,
		collectionTableName: DefaultCollectionStoreTableName,
		distanceMetric:      DefaultDistanceMetric,
	}

	for _, opt := range opts {
		opt(o)
	}

	if o.db == nil && !o.connURLSet {
		return Store{}, fmt.Errorf("%w: missing duckdb connection", ErrInvalidOptions)
	}

	if o.embedder == nil {
		return Store{}, fmt.Errorf("%w: missing embedder", ErrInvalidOptions)
	}

	if o.vectorDimensions <= 0 {
		return Store{}, fmt.Errorf("%w: vector dimensions must be > 0", ErrInvalidOptions)
	}

	validMetrics := map[string]bool{
		"cosine": true,
		"l2sq":   true,
		"ip":     true,
	}
	if !validMetrics[o.distanceMetric] {
		return Store{}, fmt.Errorf("%w: invalid distance metric %q", ErrInvalidOptions, o.distanceMetric)
	}

	return *o, nil
}
