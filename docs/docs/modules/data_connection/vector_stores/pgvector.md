---
title: "Getting Started: pgvector"
draft: true
---

# Getting Started: pgvector

[PGVector](https://github.com/pgvector/pgvector) is an open-source vector similarity search for Postgres

PGVector supports:
* exact and approximate nearest neighbor search
* L2 distance, inner product, and cosine distance
* IVFFlat and HNSW index types

See the [installation instructions](https://github.com/pgvector/pgvector#installation-notes).

## Usage with LangChainGo

In code, create an embedder based on an LLM (OpenAI, Ollama, etc.):
```go
	llm, _:= openai.New()
	emb, _ := embeddings.NewEmbedder(llm)
```

For OpenAI embeddings, you will need obtain an API key and provide as an environment variable to the program:

```bash
   export OPENAI_API_KEY=your_openai_api_key_here
```

Create a vector store:
```go
	ctx := context.Background()
	store, err := pgvector.New(
		ctx,
		pgvector.WithConnectionURL("postgres://testuser:testpass@localhost:5432/testdb?sslmode=disable"),
		pgvector.WithEmbedder(emb),
	)
```

Document tables will be created automatically.

Add documents:
```go
	_, err = store.AddDocuments(context.Background(), []schema.Document{
		{
			PageContent: "Tokyo",
			Metadata: map[string]any{
				"population": 38,
				"area":       2190,
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
```

Run a similarity search using cosine distance (`<=>`):

```go
	filter := map[string]any{"area": "1523"}

	docs, err = store.SimilaritySearch(ctx, "only cities in south america",
		10,
		vectorstores.WithScoreThreshold(0.80),
		vectorstores.WithFilters(filter),
	)
```

For now, pgvector integration only supports simple key-value filters and cosine distance search.

## Full example

Here is the entire program (from [pgvector-vectorstore-example](https://github.com/tmc/langchaingo/blob/main/examples/pgvector-vectorstore-example/pgvector_vectorstore_example.go)):
```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/vectorstores/pgvector"

	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/vectorstores"
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

	// Create a new pgvector store.
	ctx := context.Background()
	store, err := pgvector.New(
		ctx,
		pgvector.WithConnectionURL("postgres://testuser:testpass@localhost:5432/testdb?sslmode=disable"),
		pgvector.WithEmbedder(e),
	)
	if err != nil {
		log.Fatal(err)
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
		log.Fatal(err)
	}

	// Search for similar documents.
	docs, err := store.SimilaritySearch(ctx, "japan", 1)
	fmt.Println(docs)

	// Search for similar documents using score threshold.
	docs, err = store.SimilaritySearch(ctx, "only cities in south america", 10, vectorstores.WithScoreThreshold(0.80))
	fmt.Println(docs)

	// Search for similar documents using score threshold and metadata filter.
	// Metadata filter for pgvector only supports key-value pairs for now.
	filter := map[string]any{"area": "1523"} // Sao Paulo

	docs, err = store.SimilaritySearch(ctx, "only cities in south america",
		10,
		vectorstores.WithScoreThreshold(0.80),
		vectorstores.WithFilters(filter),
	)
	fmt.Println(docs)
}
```
