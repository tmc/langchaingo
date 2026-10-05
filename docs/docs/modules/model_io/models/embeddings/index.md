---
title: "Getting Started: Embeddings"
---

# Getting Started: Embeddings

<Info>
[Conceptual Guide](https://python.langchain.com/docs/modules/data_connection/text_embedding)
</Info>

Embeddings can be used to create a numerical representation of textual data. This numerical representation is useful because it can be used to find similar documents.

## Example
[`examples/vertex-embedding-example/vertex-embedding-example.go`](https://github.com/tmc/langchaingo/blob/main/examples/vertex-embedding-example/vertex-embedding-example.go):

```go
// Set the VERTEX_PROJECT to your GCP project with Vertex AI APIs enabled.
// Set VERTEX_LOCATION to a GCP location (region); if you're not sure about
// the location, set us-central1
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/tmc/langchaingo/llms/googleai"
	"github.com/tmc/langchaingo/llms/googleai/vertex"
)

func main() {
	ctx := context.Background()
	project := os.Getenv("VERTEX_PROJECT")
	location := os.Getenv("VERTEX_LOCATION")
	llm, err := vertex.New(ctx, googleai.WithCloudProject(project), googleai.WithCloudLocation(location))
	if err != nil {
		log.Fatal(err)
	}

	embeddings, err := llm.CreateEmbedding(ctx, []string{"I am a human"})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(embeddings)
}
```

## Dig deeper

- [Integrations](integrations.md)
