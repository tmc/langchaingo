package main

import (
	"context"
	"fmt"
	"log"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)

func main() {
	// Initialize the OpenAI client with Deepseek model
	llm, err := openai.New(
		openai.WithModel("deepseek-reasoner"),
		openai.WithBaseURL("https://api.deepseek.com"),
	)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// Create messages for the chat
	content := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, "You are a helpful assistant that explains complex topics step by step"),
		llms.TextParts(llms.ChatMessageTypeHuman, "Explain how quantum entanglement works and why it's important for quantum computing"),
	}

	// Generate content with streaming to see both reasoning and final answer in real-time
	completion, err := llm.GenerateContent(
		ctx,
		content,
		llms.WithMaxTokens(2000),
		llms.WithTemperature(0.7),
		llms.WithStreamingReasoningFunc(func(_ context.Context, reasoningChunk []byte, chunk []byte) error {
			if len(reasoningChunk) > 0 {
				fmt.Printf("Streaming Reasoning: %s\n", string(reasoningChunk))
			}
			if len(chunk) > 0 {
				fmt.Printf("Streaming Content: %s\n", string(chunk))
			}
			return nil
		}),
	)
	if err != nil {
		log.Fatal(err)
	}

	// Access the reasoning content and final answer separately
	if len(completion.Choices) > 0 {
		choice := completion.Choices[0]
		fmt.Printf("\n\nReasoning Process:\n%s\n", choice.ReasoningContent)
		fmt.Printf("\nFinal Answer:\n%s\n", choice.Content)
	}
}
