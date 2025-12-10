package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/bedrock"
)

func main() {
	var (
		modelID      = flag.String("model", "amazon.titan-text-lite-v1", "Model ID to use")
		provider     = flag.String("provider", "", "Explicit provider (optional)")
		prompt       = flag.String("prompt", "Say hello in one word", "Prompt to send")
		awsRegion    = flag.String("region", "us-east-1", "AWS region")
		verbose      = flag.Bool("verbose", false, "Enable verbose output")
	)
	flag.Parse()

	ctx := context.Background()

	// Create Bedrock LLM options
	opts := []bedrock.Option{
		bedrock.WithModel(*modelID),
	}
	
	// Add explicit provider if specified
	if *provider != "" {
		opts = append(opts, bedrock.WithModelProvider(*provider))
		if *verbose {
			fmt.Printf("Using explicit provider: %s\n", *provider)
		}
	}

	// Create LLM instance
	llm, err := bedrock.New(opts...)
	if err != nil {
		log.Fatalf("Failed to create Bedrock LLM: %v", err)
	}

	if *verbose {
		fmt.Printf("Model ID: %s\n", *modelID)
		fmt.Printf("AWS Region: %s\n", *awsRegion)
		fmt.Printf("Prompt: %s\n", *prompt)
		fmt.Println("---")
	}

	// Test 1: Simple Call
	fmt.Println("Testing Call method:")
	response, err := llm.Call(ctx, *prompt)
	if err != nil {
		log.Printf("Error calling model: %v", err)
	} else {
		fmt.Printf("Response: %s\n", response)
	}

	// Test 2: GenerateContent with messages
	fmt.Println("\nTesting GenerateContent method:")
	messages := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeSystem,
			Parts: []llms.ContentPart{
				llms.TextPart("You are a helpful assistant."),
			},
		},
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextPart(*prompt),
			},
		},
	}

	resp, err := llm.GenerateContent(ctx, messages)
	if err != nil {
		log.Printf("Error generating content: %v", err)
	} else {
		if len(resp.Choices) > 0 && len(resp.Choices[0].Content) > 0 {
			fmt.Printf("Response: %s\n", resp.Choices[0].Content)
		}
	}
}