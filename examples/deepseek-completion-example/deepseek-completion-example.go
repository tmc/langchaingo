package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/deepseek"
	"github.com/tmc/langchaingo/llms/openai"
)

func main() {
	fmt.Println("=== DeepSeek Integration Examples ===\n")

	// Example 1: Using the dedicated DeepSeek package (recommended)
	fmt.Println("1. Using DeepSeek package:")
	exampleWithDeepSeekPackage()

	fmt.Println("\n" + strings.Repeat("=", 50) + "\n")

	// Example 2: Using OpenAI client directly (also valid)
	fmt.Println("2. Using OpenAI client directly:")
	exampleWithOpenAIClient()
}

func exampleWithDeepSeekPackage() {
	// Initialize using the dedicated DeepSeek package
	llm, err := deepseek.New(
		deepseek.WithModel(deepseek.ModelReasoner),
	)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// Simple chat example
	fmt.Println("Simple chat:")
	response, err := llm.Chat(ctx, "What is 2+2?")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Answer: %s\n\n", response)

	// Chat with reasoning
	fmt.Println("Chat with reasoning:")
	reasoning, answer, err := llm.ChatWithReasoning(
		ctx,
		"Explain why the sky is blue",
		llms.WithMaxTokens(1000),
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Reasoning: %s\n", reasoning)
	fmt.Printf("Answer: %s\n", answer)
}

func exampleWithOpenAIClient() {
	// Initialize the OpenAI client with Deepseek model
	llm, err := openai.New(
		openai.WithModel("deepseek-reasoner"),
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
