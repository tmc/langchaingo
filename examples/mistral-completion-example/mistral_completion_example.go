package main

import (
	"context"
	"fmt"
	"log"

	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/llms/mistral"
	"github.com/vxcontrol/langchaingo/llms/streaming"
)

func main() {
	llm, err := mistral.New(mistral.WithModel("ministral-8b-latest"))
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()
	completionWithStreaming, err := llms.GenerateFromSinglePrompt(
		ctx,
		llm,
		"Who was the first man to walk on the moon?",
		llms.WithTemperature(0.8),
		llms.WithStreamingFunc(func(_ context.Context, chunk streaming.Chunk) error {
			fmt.Println(chunk.String())
			return nil
		}),
	)
	if err != nil {
		log.Fatal(err)
	}
	// The full string response will be available in completionWithStreaming after the streaming is complete.
	// (The Go compiler mandates declared variables be used at least once, hence the `_` assignment. https://go.dev/ref/spec#Blank_identifier)
	_ = completionWithStreaming

	completionWithoutStreaming, err := llms.GenerateFromSinglePrompt(ctx, llm, "Who was the first man to go to space?",
		llms.WithTemperature(0.2),
		llms.WithModel("mistral-small-latest"),
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("\n" + completionWithoutStreaming)
}
