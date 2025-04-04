package main

import (
	"context"
	"fmt"
	"log"

	"github.com/averikitsch/langchaingo/llms"
	"github.com/averikitsch/langchaingo/llms/openai"
)

func main() {
	llm, err := openai.New(openai.WithModel("gpt-4-turbo"))
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()

	content := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, "You are a company branding design wizard."),
		llms.TextParts(llms.ChatMessageTypeHuman, "What would be a good company name a company that makes colorful socks?"),
	}

	completion, err := llm.GenerateContent(ctx, content, llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
		fmt.Print(string(chunk))
		return nil
	}))
	if err != nil {
		log.Fatal(err)
	}
	_ = completion
}
