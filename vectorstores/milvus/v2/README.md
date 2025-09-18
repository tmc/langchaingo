# Milvus Vector Store v2

This package provides a Milvus vector store implementation using the new Milvus SDK v2 (`github.com/milvus-io/milvus/client/v2`).

## Background

The original `vectorstores/milvus` package uses the archived `github.com/milvus-io/milvus-sdk-go/v2` SDK, which was deprecated by the Milvus maintainers on March 21, 2025. This new `v2` package uses the actively maintained SDK at `github.com/milvus-io/milvus/client/v2`.

## Migration from v1 Package

### Quick Migration Guide

1. **Update imports**:
   ```go
   // Old
   import "github.com/tmc/langchaingo/vectorstores/milvus"

   // New
   import "github.com/tmc/langchaingo/vectorstores/milvus/v2"
   ```

2. **Update configuration** (optional - v1 configs are automatically converted):
   ```go
   // Old
   config := client.Config{Address: "localhost:19530"}

   // New (recommended)
   config := milvusclient.ClientConfig{Address: "localhost:19530"}

   // Or keep v1 config (automatically converted)
   config := client.Config{Address: "localhost:19530"} // Still works!
   ```

3. **Update function calls**:
   ```go
   // Old
   store, err := milvus.New(ctx, config, opts...)

   // New
   store, err := milvusv2.New(ctx, config, opts...)
   ```

### Compatibility Features

The v2 package includes compatibility adapters that allow gradual migration:

- **Config Compatibility**: v1 `client.Config` is automatically converted to v2 `milvusclient.ClientConfig`
- **Index Compatibility**: Use `WithIndexV1()` for v1 index types
- **Search Parameter Compatibility**: Use `WithSearchParametersV1()` for v1 search parameters
- **Metric Type Compatibility**: Use `WithMetricTypeV1()` for v1 metric types

### Example: Gradual Migration

```go
// During migration - mix v1 and v2 configurations
store, err := milvusv2.New(ctx,
    client.Config{Address: "localhost:19530"}, // v1 config
    milvusv2.WithEmbedder(embedder),
    milvusv2.WithIndexV1(oldIndex),           // v1 index
    milvusv2.WithCollectionName("my_collection"),
)
```

## Usage

### Basic Usage

```go
package main

import (
    "context"

    "github.com/milvus-io/milvus/client/v2/entity"
    "github.com/milvus-io/milvus/client/v2/index"
    "github.com/milvus-io/milvus/client/v2/milvusclient"
    "github.com/tmc/langchaingo/embeddings"
    "github.com/tmc/langchaingo/llms/openai"
    "github.com/tmc/langchaingo/schema"
    "github.com/tmc/langchaingo/vectorstores/milvus/v2"
)

func main() {
    ctx := context.Background()

    // Create embedder
    llm, err := openai.New()
    if err != nil {
        panic(err)
    }
    embedder, err := embeddings.NewEmbedder(llm)
    if err != nil {
        panic(err)
    }

    // Configure Milvus connection
    config := milvusclient.ClientConfig{
        Address: "localhost:19530",
    }

    // Create vector store
    store, err := v2.New(ctx, config,
        v2.WithEmbedder(embedder),
        v2.WithCollectionName("my_documents"),
        v2.WithIndex(index.NewAutoIndex(entity.L2)),
    )
    if err != nil {
        panic(err)
    }

    // Add documents
    docs := []schema.Document{
        {
            PageContent: "This is a document about AI",
            Metadata: map[string]any{"topic": "AI", "source": "example"},
        },
        {
            PageContent: "This document discusses machine learning",
            Metadata: map[string]any{"topic": "ML", "source": "example"},
        },
    }

    ids, err := store.AddDocuments(ctx, docs)
    if err != nil {
        panic(err)
    }

    // Search for similar documents
    results, err := store.SimilaritySearch(ctx, "artificial intelligence", 5)
    if err != nil {
        panic(err)
    }

    for _, doc := range results {
        fmt.Printf("Score: %.3f, Content: %s\n", doc.Score, doc.PageContent)
    }
}
```

### Configuration Options

```go
store, err := v2.New(ctx, config,
    v2.WithEmbedder(embedder),                    // Required: embedder for vector generation
    v2.WithCollectionName("my_collection"),       // Collection name
    v2.WithPartitionName("my_partition"),         // Partition name (optional)
    v2.WithTextField("content"),                  // Text field name (default: "text")
    v2.WithMetaField("metadata"),                 // Metadata field name (default: "meta")
    v2.WithVectorField("embedding"),              // Vector field name (default: "vector")
    v2.WithPrimaryField("id"),                    // Primary field name (default: "pk")
    v2.WithMaxTextLength(1000),                   // Max text length (default: 65535)
    v2.WithShards(2),                             // Number of shards (default: 1)
    v2.WithIndex(index.NewAutoIndex(entity.L2)),  // Vector index configuration
    v2.WithMetricType(entity.L2),                 // Distance metric
    v2.WithDropOld(),                             // Drop existing collection
    v2.WithSkipFlushOnWrite(),                    // Skip flushing after writes
)
```

## API Reference

### Core Methods

- `New(ctx, config, opts...)` - Create new Milvus vector store
- `AddDocuments(ctx, docs, opts...)` - Add documents to the store
- `SimilaritySearch(ctx, query, numDocs, opts...)` - Search for similar documents

### Configuration Options

#### Basic Options
- `WithEmbedder(embedder)` - Set the embedder (required)
- `WithCollectionName(name)` - Set collection name
- `WithPartitionName(name)` - Set partition name

#### Field Configuration
- `WithTextField(name)` - Set text field name
- `WithMetaField(name)` - Set metadata field name
- `WithVectorField(name)` - Set vector field name
- `WithPrimaryField(name)` - Set primary key field name

#### Performance Options
- `WithMaxTextLength(length)` - Set maximum text length
- `WithShards(num)` - Set number of shards
- `WithSkipFlushOnWrite()` - Skip flushing after writes

#### Index and Metrics
- `WithIndex(index)` - Set vector index
- `WithMetricType(metric)` - Set distance metric
- `WithEF(ef)` - Set EF parameter for HNSW

#### Compatibility Options (for migration)
- `WithIndexV1(index)` - Use v1 index type
- `WithSearchParametersV1(params)` - Use v1 search parameters
- `WithMetricTypeV1(metric)` - Use v1 metric type
- `WithConsistencyLevelV1(level)` - Use v1 consistency level

## Index Types

The v2 package supports the new index types from the Milvus SDK v2:

```go
// Auto index (recommended for most use cases)
index.NewAutoIndex(entity.L2)

// Flat index
index.NewFlatIndex(entity.L2)

// IVF Flat index
index.NewIvfFlatIndex(entity.L2, 1024) // metric, nlist

// HNSW index
index.NewHNSWIndex(entity.L2, 16, 200) // metric, M, efConstruction
```

## Metric Types

Supported distance metrics:

- `entity.L2` - L2 (Euclidean) distance
- `entity.IP` - Inner product
- `entity.COSINE` - Cosine similarity
- `entity.HAMMING` - Hamming distance (for binary vectors)
- `entity.JACCARD` - Jaccard distance (for binary vectors)

## Error Handling

The package defines specific error types:

- `ErrEmbedderWrongNumberVectors` - Vector count mismatch
- `ErrColumnNotFound` - Missing required column
- `ErrInvalidFilters` - Invalid filter format
- `ErrInvalidOptions` - Invalid configuration options

## Testing

The package includes comprehensive tests that cover:

- Configuration compatibility (v1/v2)
- Index compatibility
- Option functions
- Document operations
- Search operations

Run tests with:

```bash
go test ./vectorstores/milvus/v2/...
```

## Migration Timeline

- **Current**: Both packages are available
- **Recommended**: Use v2 package for new projects
- **Migration**: Use compatibility options for gradual migration
- **Future**: v1 package will be fully deprecated when Milvus 3.0 is released

## See Also

- [Migration Examples](example_migration.go) - Detailed migration examples
- [Milvus SDK v2 Documentation](https://milvus.io/docs)
- [LangChain Go Documentation](https://github.com/tmc/langchaingo)