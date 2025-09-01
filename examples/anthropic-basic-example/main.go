package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/anthropic"
)

func main() {
	if os.Getenv("ANTHROPIC_API_KEY") == "" {
		log.Fatal("ANTHROPIC_API_KEY environment variable is required")
	}

	ctx := context.Background()

	// Create Anthropic LLM
	llm, err := anthropic.New()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("=== Anthropic API Demo ===\n")

	// Example 1: Basic completion
	fmt.Println("Example 1: Basic Completion")
	fmt.Println("----------------------------")
	
	messages := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextPart("What is the capital of France?"),
			},
		},
	}

	resp, err := llm.GenerateContent(ctx, messages,
		llms.WithMaxTokens(100),
	)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	fmt.Printf("Response: %s\n", resp.Choices[0].Content)
	printTokenUsage(resp.Choices[0].GenerationInfo)

	// Example 2: Complex reasoning
	fmt.Println("\nExample 2: Complex Reasoning")
	fmt.Println("-----------------------------")
	
	messages2 := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextPart("Analyze this logic puzzle: Three boxes are labeled 'Apples', 'Oranges', and 'Apples and Oranges'. Each label is wrong. You can pick one fruit from one box. How do you figure out what's in each box?"),
			},
		},
	}

	resp2, err := llm.GenerateContent(ctx, messages2,
		llms.WithMaxTokens(500),
		llms.WithModel("claude-3-5-sonnet-20241022"),
	)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	fmt.Printf("Response: %s\n", resp2.Choices[0].Content)
	printTokenUsage(resp2.Choices[0].GenerationInfo)

	// Example 3: System design
	fmt.Println("\nExample 3: System Design")
	fmt.Println("------------------------")
	
	messages3 := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeSystem,
			Parts: []llms.ContentPart{
				llms.TextPart("You are an expert system architect with deep knowledge of distributed systems."),
			},
		},
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextPart("Design a simple URL shortener service that can handle 1000 requests per second."),
			},
		},
	}

	resp3, err := llm.GenerateContent(ctx, messages3,
		llms.WithMaxTokens(800),
		llms.WithModel("claude-3-5-sonnet-20241022"),
	)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	fmt.Printf("Response: %s\n", resp3.Choices[0].Content)
	printTokenUsage(resp3.Choices[0].GenerationInfo)

	fmt.Println("\n=== Demo Complete ===")
}

func printTokenUsage(generationInfo map[string]any) {
	fmt.Println("\nToken Usage:")
	fmt.Println("------------")
	
	// Standard tokens
	if inputTokens, ok := generationInfo["InputTokens"].(int); ok {
		fmt.Printf("Input Tokens: %d\n", inputTokens)
	}
	
	if outputTokens, ok := generationInfo["OutputTokens"].(int); ok {
		fmt.Printf("Output Tokens: %d\n", outputTokens)
	}
	
	fmt.Println()
}