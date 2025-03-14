package alloydb

import (
	"errors"

	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/internal/alloydbutil"
	"github.com/tmc/langchaingo/vectorstores"
)

const (
	defaultSchemaName         = "public"
	defaultIDColumn           = "langchain_id"
	defaultContentColumn      = "content"
	defaultEmbeddingColumn    = "embedding"
	defaultMetadataJsonColumn = "langchain_metadata"
	defaultK                  = 4
)

// AlloyDBVectoreStoresOption is a function for creating new vector store
// with other than the default values.
type AlloyDBVectoreStoresOption func(vs *VectorStore)

// WithSchemaName sets the VectorStore's schemaName field.
func WithSchemaName(schemaName string) AlloyDBVectoreStoresOption {
	return func(v *VectorStore) {
		v.schemaName = schemaName
	}
}

// WithContentColumn sets VectorStore's the idColumn field.
func WithIDColumn(idColumn string) AlloyDBVectoreStoresOption {
	return func(v *VectorStore) {
		v.idColumn = idColumn
	}
}

// WithMetadataJsonColumn sets VectorStore's the metadataJsonColumn field.
func WithMetadataJsonColumn(metadataJsonColumn string) AlloyDBVectoreStoresOption {
	return func(v *VectorStore) {
		v.metadataJsonColumn = metadataJsonColumn
	}
}

// WithContentColumn sets the VectorStore's ContentColumn field.
func WithContentColumn(contentColumn string) AlloyDBVectoreStoresOption {
	return func(v *VectorStore) {
		v.contentColumn = contentColumn
	}
}

// WithEmbeddingColumn sets the EmbeddingColumn field.
func WithEmbeddingColumn(embeddingColumn string) AlloyDBVectoreStoresOption {
	return func(v *VectorStore) {
		v.embeddingColumn = embeddingColumn
	}
}

// WithMetadataColumns sets the VectorStore's MetadataColumns field.
func WithMetadataColumns(metadataColumns []string) AlloyDBVectoreStoresOption {
	return func(v *VectorStore) {
		v.metadataColumns = metadataColumns
	}
}

// WithK sets the number of Documents to return from the VectorStore.
func WithK(k int) AlloyDBVectoreStoresOption {
	return func(v *VectorStore) {
		v.k = k
	}
}

// WithDistanceStrategy sets the distance strategy used by the VectorStore.
func WithDistanceStrategy(distanceStrategy distanceStrategy) AlloyDBVectoreStoresOption {
	return func(v *VectorStore) {
		v.distanceStrategy = distanceStrategy
	}
}

// applyAlloyDBVectorStoreOptions applies the given VectorStore options to the
// VectorStore with an alloydb Engine.
func applyAlloyDBVectorStoreOptions(engine alloydbutil.PostgresEngine, embedder embeddings.Embedder, tableName string, opts ...AlloyDBVectoreStoresOption) (VectorStore, error) {
	// Check for required values.
	if engine.Pool == nil {
		return VectorStore{}, errors.New("missing vector store engine")
	}
	if embedder == nil {
		return VectorStore{}, errors.New("missing vector store embeder")
	}
	if tableName == "" {
		return VectorStore{}, errors.New("missing vector store table name")
	}
	vs := &VectorStore{
		engine:             engine,
		embedder:           embedder,
		tableName:          tableName,
		schemaName:         defaultSchemaName,
		idColumn:           defaultIDColumn,
		contentColumn:      defaultContentColumn,
		embeddingColumn:    defaultEmbeddingColumn,
		metadataJsonColumn: defaultMetadataJsonColumn,
		k:                  defaultK,
		distanceStrategy:   defaultDistanceStrategy,
		metadataColumns:    []string{},
	}
	for _, opt := range opts {
		opt(vs)
	}

	return *vs, nil
}

func applyOpts(options ...vectorstores.Option) (vectorstores.Options, error) {
	opts := vectorstores.Options{}
	for _, opt := range options {
		opt(&opts)
	}
	return opts, nil
}
