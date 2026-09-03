package main

import (
	"context"
	"fmt"
	"log"

	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/llms/openai"
	"github.com/vxcontrol/langchaingo/llms/streaming"
)

func main() {
	// Initialize the OpenAI client with Deepseek model
	// Per the official DeepSeek API documentation (https://api-docs.deepseek.com/),
	// please note that all API calls must be configured to use the DeepSeek base URL instead of the OpenAI endpoint.
	llm, err := openai.New(
		openai.WithModel("deepseek-v4-flash"),
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
		llms.WithStreamingFunc(func(_ context.Context, chunk streaming.Chunk) error {
			if chunk.Type == streaming.ChunkTypeReasoning {
				if !chunk.Reasoning.IsEmpty() {
					fmt.Printf("Streaming Reasoning: %s\n", chunk.Reasoning.Content)
				} else {
					fmt.Println("Streaming Reasoning chunk is nil")
				}
			}
			if chunk.Type == streaming.ChunkTypeText {
				fmt.Printf("Streaming Content: %s\n", chunk.Content)
			}
			if chunk.Type == streaming.ChunkTypeDone {
				fmt.Println("Streaming Done")
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
		fmt.Printf("\n\nReasoning Process:\n%s\n", choice.Reasoning.Content)
		fmt.Printf("\nFinal Answer:\n%s\n", choice.Content)
	}
}
