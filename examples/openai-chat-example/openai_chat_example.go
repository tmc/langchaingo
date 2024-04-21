package main

import (
	"context"
	"fmt"
	"log"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)

func main() {
	llm, err := openai.New()
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()

	content := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, "You are a company branding design wizard."),
		llms.TextParts(llms.ChatMessageTypeHuman, "What would be a good company name a company that makes colorful socks?"),
	}

	if _, err := llm.GenerateContent(ctx, content,
		llms.WithMaxTokens(1024),
		llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
			fmt.Print(string(chunk))
			return nil
		})); err != nil {
		log.Fatal(err)
	}
}
