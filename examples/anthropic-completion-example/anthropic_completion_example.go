package main

import (
	"context"
	"fmt"
	"log"

	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/llms/anthropic"
	"github.com/vxcontrol/langchaingo/llms/streaming"
)

func main() {
	llm, err := anthropic.New(
		anthropic.WithModel("claude-sonnet-4-5"),
	)
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()
	completion, err := llms.GenerateFromSinglePrompt(ctx, llm, "Hi claude, write a poem about golang powered AI systems",
		llms.WithTemperature(0.8),
		llms.WithStreamingFunc(func(_ context.Context, chunk streaming.Chunk) error {
			fmt.Println(chunk.String())
			return nil
		}),
	)
	if err != nil {
		log.Fatal(err)
	}

	_ = completion
}
