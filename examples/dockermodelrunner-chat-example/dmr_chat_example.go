package main

import (
	"context"
	"fmt"
	"log"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)

func main() {
	opts := []openai.Option{
		openai.WithBaseURL("http://localhost:12434/engines/v1"),
		openai.WithModel("ai/llama3.2:latest"),
	}

	llm, err := openai.New(opts...)
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()

	content := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, "You are a company branding design wizard."),
		llms.TextParts(llms.ChatMessageTypeHuman, "What would be a good company name a company that makes colorful clothes for whales?"),
	}

	if _, err := llm.GenerateContent(ctx, content,
		llms.WithMaxTokens(1024),
		llms.WithTemperature(0.8),
		llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
			fmt.Print(string(chunk))
			return nil
		})); err != nil {
		log.Fatal(err)
	}
}
