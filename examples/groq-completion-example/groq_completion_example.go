package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/vendasta/langchaingo/llms"
	"github.com/vendasta/langchaingo/llms/openai"
)

func main() {
	apiKey := os.Getenv("GROQ_API_KEY")

	llm, err := openai.New(
		openai.WithModel("moonshotai/kimi-k2-instruct"),
		openai.WithBaseURL("https://api.groq.com/openai/v1"),
		openai.WithToken(apiKey),
	)
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()
	_, err = llms.GenerateFromSinglePrompt(ctx,
		llm,
		"Write a long poem about how golang is a fantastic language.",
		llms.WithTemperature(0.8),
		llms.WithMaxTokens(4096),
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
