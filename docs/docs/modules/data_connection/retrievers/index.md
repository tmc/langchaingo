---
title: Retrievers
---

# Retrievers

<Info>
[Conceptual Guide](https://python.langchain.com/docs/modules/data_connection/retrievers/)
</Info>
A retriever returns documents relevant to a query. LangChainGo's
`schema.Retriever` interface exposes `GetRelevantDocuments`, which accepts
a context and query string and returns documents or an error.

A vector store can be adapted to this interface with
`vectorstores.ToRetriever`. The example below uses Ollama for embeddings
and Qdrant to search a collection.

## Configure an embedder

```go
	llm, err := ollama.New(ollama.WithModel("llama2"))

	if err != nil {
		log.Fatal(err)
	}

	embedder, err := embeddings.NewEmbedder(llm)
	if err != nil {
		log.Fatal(err)
	}
```

## Configure a vector store

Connect to Qdrant and select the collection to search:
```go
	url, err := url.Parse("http://localhost:6333")
	if err != nil {
		log.Fatal(err)
	}

	store, err := qdrant.New(
		qdrant.WithURL(*url),
		qdrant.WithCollectionName("youtube_transcript"),
		qdrant.WithEmbedder(embedder),
	)
	if err != nil {
		log.Fatal(err)
	}

```

## Retrieve documents

Adapt the store and search for relevant documents:
```go
	searchQuery := "how to make a cake"
	
	// Create retriever with basic configuration
	retriever := vectorstores.ToRetriever(store, 10)
	
	// Search for relevant documents
	resDocs, err := retriever.GetRelevantDocuments(context.Background(), searchQuery)
	if err != nil {
		log.Fatal(err)
	}
```

See the [Qdrant example](https://github.com/tmc/langchaingo/tree/main/examples/qdrant-vectorstore-example) for a complete program.
