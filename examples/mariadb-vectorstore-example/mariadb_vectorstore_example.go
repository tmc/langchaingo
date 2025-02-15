package main

import (
	"context"
	"fmt"
	"log"

	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/vectorstores"
	"github.com/tmc/langchaingo/vectorstores/mariadb"
)

func main() {
	llm, err := openai.New()
	if err != nil {
		log.Fatal(err)
	}
	e, err := embeddings.NewEmbedder(llm)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	store, err := mariadb.New(
		ctx,
		mariadb.WithDSN("root:test@tcp(localhost:3306)/test"),
		mariadb.WithEmbedder(e),
		mariadb.WithPreDeleteCollection(true),
	)
	if err != nil {
		log.Fatal(err)
	}
	docs := []schema.Document{
		{
			PageContent: "Vladivostok",
			Metadata: map[string]any{
				"Country":    "Russia",
				"Population": 600000,
			},
		},
		{
			PageContent: "Moscow",
			Metadata: map[string]any{
				"Country":    "Russia",
				"Population": 12500000,
			},
		},
		{
			PageContent: "New York",
			Metadata: map[string]any{
				"Country":    "USA",
				"Population": 8500000,
			},
		},
		{
			PageContent: "London",
			Metadata: map[string]any{
				"Country":    "England",
				"Population": 9000000,
			},
		},
	}
	_, err = store.AddDocuments(ctx, docs)
	if err != nil {
		log.Fatal(err)
	}

	// Search for similar documents.
	docs, err = store.SimilaritySearch(ctx, "England", 1)
	fmt.Println("SimularitySearch for England:", docs)

	// Search for similar documents using score threshold.
	docs, err = store.SimilaritySearch(ctx, "only cities in russia", 10, vectorstores.WithScoreThreshold(0.80))
	fmt.Println("SimularitySearch with Threshold:", docs)

	// Using filters.
	filter := map[string]any{"Country": "USA"}
	docs, err = store.SimilaritySearch(ctx, "Cities",
		1,
		vectorstores.WithScoreThreshold(0.80),
		vectorstores.WithFilters(filter),
	)
	fmt.Println("Filter results:", docs)

	// Using advanced filters.
	advancedFilter := map[string]any{
		"Country !=":   "Russia",
		"Population >": 8500000,
	}
	docs, err = store.Search(ctx, advancedFilter, 5)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Advanced filter results:", docs)
}
