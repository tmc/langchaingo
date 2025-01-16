package main

import (
	"context"
	"fmt"
	"log"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)

func main() {
	llm, err := openai.New(openai.WithModel("gpt-4o"))
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()

	content := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, "You are a Birds expert"),
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.ImageURLContent{
					URL:    "https://upload.wikimedia.org/wikipedia/commons/3/39/Brown-Falcon%2C-Vic%2C-3.1.2008.jpg",
					Detail: "low",
				},
				llms.TextPart("describe this image in detail, what bird is it?"),
			},
		},
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
