package main

import (
	"context"
	"fmt"
	"log"
	"net/url"

	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/vectorstores"
	"github.com/tmc/langchaingo/vectorstores/vearch"
)

func main() {
		// Create an embeddings client using the OpenAI API. Requires environment variable OPENAI_API_KEY to be set.
		llm, err := openai.New(openai.WithEmbeddingModel("model/bge-small-en-v1.5"))// Specify your preferred embedding model
		if err != nil {
				log.Fatal(err)
		}

		e, err := embeddings.NewEmbedder(llm)
		if err != nil {
			log.Fatal(err)
		}

		ctx := context.Background()

		// Create a new Vearch vector store.
		store, err := vearch.New(
			vearch.WithDbName("langchaingo_dbt"),
			vearch.WithSpaceName("langchaingo_t"),
			vearch.WithURL("your router url"),
			vearch.WithEmbedder(e),
		)
		if err != nil {
			log.Fatal(err)
		}

		// Add documents to the Vearch vector store.
		var ids []string
		ids, err = store.AddDocuments(context.Background(), []schema.Document{
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
					"population": 95,
					"area":       1572,
				},
			},
		})
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("ids",ids)

		// Search for similar documents.
		docs, err := store.SimilaritySearch(ctx, "japan", 1)
		fmt.Println(docs)

		// Search for similar documents usingmetadata filter.
		filter := map[string]interface{}{
			"AND": []map[string]interface{}{
				{ 
					"condition": map[string]interface{}{
						"Field": "population",
						"Operator":">",
						"Value": 20,
						},
				},
				},
			}
		var docs_f []schema.Document 
		docs_f, err = store.SimilaritySearch(ctx, "only cities in earth",
			10,
			vectorstores.WithFilters(filter))
		fmt.Println(docs_f)
	}