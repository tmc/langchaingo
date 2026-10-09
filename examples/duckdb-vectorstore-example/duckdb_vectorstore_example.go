package main

import (
	"context"
	"fmt"
	"log"

	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/vectorstores"
	duckdbstore "github.com/tmc/langchaingo/vectorstores/duckdb"
)

func main() {
	// Create an embeddings client using the OpenAI API. Requires environment variable OPENAI_API_KEY to be set.
	llm, err := openai.New()
	if err != nil {
		log.Fatal(err)
	}

	e, err := embeddings.NewEmbedder(llm)
	if err != nil {
		log.Fatal(err)
	}

	// Create a new DuckDB store. An empty connection URL creates an in-memory database.
	ctx := context.Background()
	store, err := duckdbstore.New(
		ctx,
		duckdbstore.WithConnectionURL(""),
		duckdbstore.WithEmbedder(e),
		duckdbstore.WithVectorDimensions(1536),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer store.Close()

	// Add documents to the DuckDB store.
	_, err = store.AddDocuments(ctx, []schema.Document{
		{
			PageContent: "Tokyo",
			Metadata: map[string]any{
				"population": 38,
				"area":       2190,
			},
		},
		{
			PageContent: "Paris",
			Metadata: map[string]any{
				"population": 11,
				"area":       105,
			},
		},
		{
			PageContent: "London",
			Metadata: map[string]any{
				"population": 9.5,
				"area":       1572,
			},
		},
		{
			PageContent: "Santiago",
			Metadata: map[string]any{
				"population": 6.9,
				"area":       641,
			},
		},
		{
			PageContent: "Buenos Aires",
			Metadata: map[string]any{
				"population": 15.5,
				"area":       203,
			},
		},
		{
			PageContent: "Rio de Janeiro",
			Metadata: map[string]any{
				"population": 13.7,
				"area":       1200,
			},
		},
		{
			PageContent: "Sao Paulo",
			Metadata: map[string]any{
				"population": 22.6,
				"area":       1523,
			},
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	// Search for similar documents.
	docs, err := store.SimilaritySearch(ctx, "japan", 1)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(docs)

	// Search for similar documents using score threshold.
	docs, err = store.SimilaritySearch(ctx, "only cities in south america", 10, vectorstores.WithScoreThreshold(0.80))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(docs)

	// Search for similar documents using score threshold and metadata filter.
	// Note: json_extract_string converts numeric metadata values to strings.
	filter := map[string]any{"area": 1523} // Sao Paulo

	docs, err = store.SimilaritySearch(ctx, "only cities in south america",
		10,
		vectorstores.WithScoreThreshold(0.80),
		vectorstores.WithFilters(filter),
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(docs)
}
