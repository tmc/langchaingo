package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/llms/openai"
	"github.com/vxcontrol/langchaingo/llms/streaming"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	apiKey := os.Getenv("PERPLEXITY_API_KEY")

	llm, err := openai.New(
		// Supported models: https://docs.perplexity.ai/docs/model-cards
		openai.WithModel("sonar-pro"),
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
		llms.WithStreamingFunc(func(_ context.Context, chunk streaming.Chunk) error {
			fmt.Println(chunk.String())
			return nil
		}),
	)
	fmt.Println()
	if err != nil {
		log.Fatal(err)
	}
}
