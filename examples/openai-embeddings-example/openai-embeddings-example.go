package main

import (
	"context"
	"fmt"
	"log"

	"github.com/averikitsch/langchaingo/llms/openai"
)

func main() {
	opts := []openai.Option{
		openai.WithModel("gpt-3.5-turbo-0125"),
		openai.WithEmbeddingModel("text-embedding-3-large"),
	}
	llm, err := openai.New(opts...)
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()
	embedings, err := llm.CreateEmbedding(ctx, []string{"ola", "mundo"})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(embedings)
}
