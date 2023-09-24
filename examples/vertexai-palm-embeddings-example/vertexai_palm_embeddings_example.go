package main

import (
	"context"
	"fmt"
	"log"

	"github.com/tmc/langchaingo/llms/vertexai"
)

func main() {
	llm, err := vertexai.NewChat()
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()
	embeddings, err := llm.CreateEmbedding(ctx, []string{"I am a human"})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(embeddings)
}
