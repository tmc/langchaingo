package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"
	
	"github.com/yincongcyincong/langchaingo/embeddings"
	"github.com/yincongcyincong/langchaingo/llms/mistral"
	"github.com/yincongcyincong/langchaingo/schema"
	"github.com/yincongcyincong/langchaingo/vectorstores"
	"github.com/yincongcyincong/langchaingo/vectorstores/pgvector"
)

func main() {
	var dsn string
	flag.StringVar(&dsn, "dsn", "", "PGvector connection string")
	flag.Parse()
	model, err := mistral.New()
	if err != nil {
		panic(err)
	}
	
	e, err := embeddings.NewEmbedder(model)
	
	if err != nil {
		panic(err)
	}
	
	// Create a new pgvector store.
	ctx := context.Background()
	store, err := pgvector.New(
		ctx,
		pgvector.WithConnectionURL(dsn),
		pgvector.WithEmbedder(e),
	)
	if err != nil {
		log.Fatal("pgvector.New", err)
	}
	
	// Add documents to the pgvector store.
	_, err = store.AddDocuments(context.Background(), []schema.Document{
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
		log.Fatal("store.AddDocuments:\n", err)
	}
	time.Sleep(1 * time.Second)
	
	// Search for similar documents.
	docs, err := store.SimilaritySearch(ctx, "japan", 1)
	if err != nil {
		log.Fatal("store.SimilaritySearch1:\n", err)
	}
	fmt.Println("store.SimilaritySearch1:\n", docs)
	
	time.Sleep(2 * time.Second) // Don't trigger cloudflare
	
	// Search for similar documents using score threshold.
	docs, err = store.SimilaritySearch(ctx, "only cities in south america", 3, vectorstores.WithScoreThreshold(0.50))
	if err != nil {
		log.Fatal("store.SimilaritySearch2:\n", err)
	}
	fmt.Println("store.SimilaritySearch2:\n", docs)
	
	time.Sleep(3 * time.Second) // Don't trigger cloudflare
	
	// Search for similar documents using score threshold and metadata filter.
	// Metadata filter for pgvector only supports key-value pairs for now.
	filter := map[string]any{"area": "1523"} // Sao Paulo
	
	docs, err = store.SimilaritySearch(ctx, "only cities in south america",
		3,
		vectorstores.WithScoreThreshold(0.50),
		vectorstores.WithFilters(filter),
	)
	if err != nil {
		log.Fatal("store.SimilaritySearch3:\n", err)
	}
	fmt.Println("store.SimilaritySearch3:\n", docs)
}
