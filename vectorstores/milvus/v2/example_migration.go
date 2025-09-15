// Package example demonstrates migration from milvus v1 package to v2 package.
//
//go:build ignore

// This file provides examples of how to migrate existing code from the deprecated
// vectorstores/milvus package to the new vectorstores/milvus/v2 package.
package main

import (
	"context"
	"fmt"
	"log"

	// Old SDK imports (deprecated)
	oldclient "github.com/milvus-io/milvus-sdk-go/v2/client"
	oldentity "github.com/milvus-io/milvus-sdk-go/v2/entity"

	// New SDK imports
	"github.com/milvus-io/milvus/client/v2/entity"
	"github.com/milvus-io/milvus/client/v2/index"
	"github.com/milvus-io/milvus/client/v2/milvusclient"

	// LangChain Go packages
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/schema"
	oldmilvus "github.com/tmc/langchaingo/vectorstores/milvus"
	newmilvus "github.com/tmc/langchaingo/vectorstores/milvus/v2"
)

// MigrationExample demonstrates how to migrate from v1 to v2.
func main() {
	ctx := context.Background()

	// Create embedder (same for both versions)
	llm, err := openai.New()
	if err != nil {
		log.Fatal(err)
	}
	embedder, err := embeddings.NewEmbedder(llm)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("=== Migration Example: v1 to v2 ===")

	// Example 1: Basic migration
	fmt.Println("\n1. Basic Configuration Migration:")
	demonstrateBasicMigration(ctx, embedder)

	// Example 2: Index migration
	fmt.Println("\n2. Index Configuration Migration:")
	demonstrateIndexMigration(ctx, embedder)

	// Example 3: Search parameter migration
	fmt.Println("\n3. Search Parameter Migration:")
	demonstrateSearchParamMigration(ctx, embedder)

	// Example 4: Compatibility layer usage
	fmt.Println("\n4. Compatibility Layer Usage:")
	demonstrateCompatibilityLayer(ctx, embedder)
}

// demonstrateBasicMigration shows how basic configuration migrates from v1 to v2.
func demonstrateBasicMigration(ctx context.Context, embedder *embeddings.EmbedderImpl) {
	// OLD WAY (v1 - deprecated)
	fmt.Println("OLD (v1):")
	fmt.Println(`
	config := client.Config{
		Address: "localhost:19530",
	}
	store, err := milvus.New(ctx, config,
		milvus.WithEmbedder(embedder),
		milvus.WithCollectionName("my_collection"),
	)`)

	// NEW WAY (v2)
	fmt.Println("\nNEW (v2):")
	fmt.Println(`
	config := milvusclient.ClientConfig{
		Address: "localhost:19530",
	}
	store, err := milvusv2.New(ctx, config,
		milvusv2.WithEmbedder(embedder),
		milvusv2.WithCollectionName("my_collection"),
	)`)

	// COMPATIBILITY WAY (v2 with v1 config)
	fmt.Println("\nCOMPATIBILITY (v2 with v1 config):")
	fmt.Println(`
	// v1 config still works with v2 package via adapter
	oldConfig := client.Config{
		Address: "localhost:19530",
	}
	store, err := milvusv2.New(ctx, oldConfig,
		milvusv2.WithEmbedder(embedder),
		milvusv2.WithCollectionName("my_collection"),
	)`)
}

// demonstrateIndexMigration shows how index configuration migrates.
func demonstrateIndexMigration(ctx context.Context, embedder *embeddings.EmbedderImpl) {
	// OLD WAY (v1)
	fmt.Println("OLD (v1):")
	fmt.Println(`
	oldIndex, err := entity.NewIndexAUTOINDEX(entity.L2)
	if err != nil {
		return err
	}
	store, err := milvus.New(ctx, config,
		milvus.WithEmbedder(embedder),
		milvus.WithIndex(oldIndex),
	)`)

	// NEW WAY (v2)
	fmt.Println("\nNEW (v2):")
	fmt.Println(`
	newIndex := index.NewAutoIndex(entity.L2)
	store, err := milvusv2.New(ctx, config,
		milvusv2.WithEmbedder(embedder),
		milvusv2.WithIndex(newIndex),
	)`)

	// COMPATIBILITY WAY
	fmt.Println("\nCOMPATIBILITY (v2 with v1 index):")
	fmt.Println(`
	// v1 index converted automatically
	oldIndex, err := oldentity.NewIndexAUTOINDEX(oldentity.L2)
	if err != nil {
		return err
	}
	store, err := milvusv2.New(ctx, config,
		milvusv2.WithEmbedder(embedder),
		milvusv2.WithIndexV1(oldIndex), // Note: WithIndexV1
	)`)
}

// demonstrateSearchParamMigration shows search parameter migration.
func demonstrateSearchParamMigration(ctx context.Context, embedder *embeddings.EmbedderImpl) {
	// OLD WAY (v1)
	fmt.Println("OLD (v1):")
	fmt.Println(`
	searchParam := entity.NewSearchParam(entity.HNSW)
	searchParam.AddRadius(0.1)
	store, err := milvus.New(ctx, config,
		milvus.WithEmbedder(embedder),
		milvus.WithSearchParameters(searchParam),
	)`)

	// NEW WAY (v2)
	fmt.Println("\nNEW (v2):")
	fmt.Println(`
	searchParams := map[string]interface{}{
		"ef": 64,
		"radius": 0.1,
	}
	store, err := milvusv2.New(ctx, config,
		milvusv2.WithEmbedder(embedder),
		milvusv2.WithSearchParameters(searchParams),
	)`)

	// COMPATIBILITY WAY
	fmt.Println("\nCOMPATIBILITY (v2 with v1 search params):")
	fmt.Println(`
	// v1 search params converted automatically
	oldSearchParam := oldentity.NewSearchParam(oldentity.HNSW)
	oldSearchParam.AddRadius(0.1)
	store, err := milvusv2.New(ctx, config,
		milvusv2.WithEmbedder(embedder),
		milvusv2.WithSearchParametersV1(oldSearchParam), // Note: WithSearchParametersV1
	)`)
}

// demonstrateCompatibilityLayer shows the compatibility layer in action.
func demonstrateCompatibilityLayer(ctx context.Context, embedder *embeddings.EmbedderImpl) {
	fmt.Println("The v2 package provides compatibility options for gradual migration:")
	fmt.Println()

	// Show all compatibility functions
	fmt.Println("Compatibility Functions Available:")
	fmt.Println("- WithIndexV1(oldentity.Index)")
	fmt.Println("- WithSearchParametersV1(oldentity.SearchParam)")
	fmt.Println("- WithMetricTypeV1(oldentity.MetricType)")
	fmt.Println("- WithConsistencyLevelV1(oldentity.ConsistencyLevel)")
	fmt.Println()

	fmt.Println("Config Compatibility:")
	fmt.Println("- client.Config (v1) -> automatically converted")
	fmt.Println("- milvusclient.ClientConfig (v2) -> used directly")
	fmt.Println("- string address -> converted to ClientConfig")
}

// exampleV1Usage shows typical v1 usage (deprecated).
func exampleV1Usage(ctx context.Context, embedder *embeddings.EmbedderImpl) error {
	// This is the old way - now deprecated
	config := oldclient.Config{
		Address: "localhost:19530",
	}

	oldIndex, err := oldentity.NewIndexAUTOINDEX(oldentity.L2)
	if err != nil {
		return err
	}

	// Using deprecated package
	store, err := oldmilvus.New(ctx, config,
		oldmilvus.WithEmbedder(embedder),
		oldmilvus.WithCollectionName("my_collection"),
		oldmilvus.WithIndex(oldIndex),
	)
	if err != nil {
		return err
	}

	// Add documents
	docs := []schema.Document{
		{PageContent: "Hello world", Metadata: map[string]any{"source": "test"}},
	}
	_, err = store.AddDocuments(ctx, docs)
	return err
}

// exampleV2Usage shows new v2 usage (recommended).
func exampleV2Usage(ctx context.Context, embedder *embeddings.EmbedderImpl) error {
	// This is the new way - recommended
	config := milvusclient.ClientConfig{
		Address: "localhost:19530",
	}

	newIndex := index.NewAutoIndex(entity.L2)

	// Using new v2 package
	store, err := newmilvus.New(ctx, config,
		newmilvus.WithEmbedder(embedder),
		newmilvus.WithCollectionName("my_collection"),
		newmilvus.WithIndex(newIndex),
	)
	if err != nil {
		return err
	}

	// Add documents (API is the same)
	docs := []schema.Document{
		{PageContent: "Hello world", Metadata: map[string]any{"source": "test"}},
	}
	_, err = store.AddDocuments(ctx, docs)
	return err
}

// exampleMixedUsage shows using v2 with v1 configurations for gradual migration.
func exampleMixedUsage(ctx context.Context, embedder *embeddings.EmbedderImpl) error {
	// Use v1 config with v2 package during migration
	oldConfig := oldclient.Config{
		Address: "localhost:19530",
	}

	oldIndex, err := oldentity.NewIndexAUTOINDEX(oldentity.L2)
	if err != nil {
		return err
	}

	// Using v2 package with v1 configuration
	store, err := newmilvus.New(ctx, oldConfig, // v1 config automatically converted
		newmilvus.WithEmbedder(embedder),
		newmilvus.WithCollectionName("my_collection"),
		newmilvus.WithIndexV1(oldIndex), // v1 index with compatibility function
		newmilvus.WithMetricTypeV1(oldentity.L2), // v1 metric type
	)
	if err != nil {
		return err
	}

	// API usage is identical
	docs := []schema.Document{
		{PageContent: "Hello world", Metadata: map[string]any{"source": "test"}},
	}
	_, err = store.AddDocuments(ctx, docs)
	return err
}