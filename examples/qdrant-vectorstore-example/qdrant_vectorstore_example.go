package main

import (
	"context"
	"fmt"
	"log"
	"net/url"
	
	"github.com/yincongcyincong/langchaingo/embeddings"
	"github.com/yincongcyincong/langchaingo/llms/openai"
	"github.com/yincongcyincong/langchaingo/schema"
	"github.com/yincongcyincong/langchaingo/vectorstores"
	"github.com/yincongcyincong/langchaingo/vectorstores/qdrant"
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
	
	ctx := context.Background()
	
	// Create a new Qdrant vector store.
	url, err := url.Parse("YOUR_QDRANT_URL")
	if err != nil {
		log.Fatal(err)
	}
	store, err := qdrant.New(
		qdrant.WithURL(*url),
		qdrant.WithCollectionName("YOUR_COLLECTION_NAME"),
		qdrant.WithEmbedder(e),
	)
	if err != nil {
		log.Fatal(err)
	}
	
	// Add documents to the Qdrant vector store.
	_, err = store.AddDocuments(context.Background(), []schema.Document{
		{
			PageContent: "A city in texas",
			Metadata: map[string]any{
				"area": 3251,
			},
		},
		{
			PageContent: "A country in Asia",
			Metadata: map[string]any{
				"area": 2342,
			},
		},
		{
			PageContent: "A country in South America",
			Metadata: map[string]any{
				"area": 432,
			},
		},
		{
			PageContent: "An island nation in the Pacific Ocean",
			Metadata: map[string]any{
				"area": 6531,
			},
		},
		{
			PageContent: "A mountainous country in Europe",
			Metadata: map[string]any{
				"area": 1211,
			},
		},
		{
			PageContent: "A lost city in the Amazon",
			Metadata: map[string]any{
				"area": 1223,
			},
		},
		{
			PageContent: "A city in England",
			Metadata: map[string]any{
				"area": 4324,
			},
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	
	// Search for similar documents.
	docs, err := store.SimilaritySearch(ctx, "england", 1)
	fmt.Println(docs)
	
	// Search for similar documents using score threshold.
	docs, err = store.SimilaritySearch(ctx, "american places", 10, vectorstores.WithScoreThreshold(0.80))
	fmt.Println(docs)
	
	// Search for similar documents using score threshold and metadata filter.
	filter := map[string]interface{}{
		"must": []map[string]interface{}{
			{
				"key": "area",
				"range": map[string]interface{}{
					"lte": 3000,
				},
			},
		},
	}
	
	docs, err = store.SimilaritySearch(ctx, "only cities in south america",
		10,
		vectorstores.WithScoreThreshold(0.80),
		vectorstores.WithFilters(filter))
	fmt.Println(docs)
}
