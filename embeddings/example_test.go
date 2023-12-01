package embeddings_test

import (
	"context"
	"log"

	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/llms/openai"
)

func Example() {
	llm, err := openai.New()
	if err != nil {
		log.Fatal(err)
	}

	embedder, err := embeddings.NewEmbedder(llm)
	if err != nil {
		log.Fatal(err)
	}

	docs := []string{"doc 1", "another doc"}
	embs, err := embedder.EmbedDocuments(context.Background(), docs)
	if err != nil {
		log.Fatal(err)
	}

	// Consume embs
	_ = embs
}
