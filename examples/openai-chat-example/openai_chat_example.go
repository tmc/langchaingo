package main

import (
	"context"
	"fmt"
	"log"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/schema"
)

func main() {
	llm, err := openai.New()
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()

	content := []schema.MessageContent{
		schema.TextParts(schema.ChatMessageTypeSystem, "You are a company branding design wizard."),
		schema.TextParts(schema.ChatMessageTypeHuman, "What would be a good company name a company that makes colorful socks?"),
	}

	completion, err := llm.GenerateContent(ctx, content, llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
		fmt.Print(string(chunk))
		return nil
	}))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(completion)
}
