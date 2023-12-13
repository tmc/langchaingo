package main

import (
	"context"
	"fmt"
	"log"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/ollama"
	"github.com/tmc/langchaingo/schema"
)

func main() {
	llm, err := ollama.NewChat(ollama.WithLLMOptions(ollama.WithModel("mistral")))
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()
	completion, err := llm.Call(ctx, []schema.ChatMessage{
		schema.SystemChatMessage{Content: "You are a hyper-creative rhyming machine."},
		schema.HumanChatMessage{Content: "Tell me a poem about Sam Altman"},
	}, llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
		fmt.Print(string(chunk))
		return nil
	}),
	)
	if err != nil {
		log.Fatal(err)
	}
	_ = completion
}
