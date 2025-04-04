package main

import (
	"context"
	"fmt"
	"log"

	"github.com/averikitsch/langchaingo/llms"
	"github.com/averikitsch/langchaingo/llms/anthropic"
)

func main() {
	llm, err := anthropic.New(
		anthropic.WithModel("claude-3-5-sonnet-20240620"),
	)
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()
	completion, err := llms.GenerateFromSinglePrompt(ctx, llm, "Hi claude, write a poem about golang powered AI systems",
		llms.WithTemperature(0.8),
		llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
			fmt.Print(string(chunk))
			return nil
		}),
	)
	if err != nil {
		log.Fatal(err)
	}

	_ = completion
}
