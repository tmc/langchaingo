package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/0xDezzy/langchaingo/llms"
	"github.com/0xDezzy/langchaingo/llms/openai"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	apiKey := os.Getenv("PERPLEXITY_API_KEY")

	llm, err := openai.New(
		// Supported models: https://docs.perplexity.ai/docs/model-cards
		openai.WithModel("llama-3.1-sonar-large-128k-online"),
		openai.WithBaseURL("https://api.perplexity.ai"),
		openai.WithToken(apiKey),
	)
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()
	_, err = llms.GenerateFromSinglePrompt(ctx,
		llm,
		"What is a prime number?",
		llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
			fmt.Print(string(chunk))
			return nil
		}),
	)
	fmt.Println()
	if err != nil {
		log.Fatal(err)
	}
}
