package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/githubmodels"
)

func main() {
	// Check if GitHub token is set
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		log.Fatal("GITHUB_TOKEN environment variable is not set")
	}

	// Create a new GitHub Models LLM
	llm, err := githubmodels.New(
		githubmodels.WithToken(token),
		githubmodels.WithModel("openai/gpt-4.1"), // Can be changed to any model available in GitHub Models
	)
	if err != nil {
		log.Fatalf("Failed to create GitHub Models LLM: %v", err)
	}

	// Create context
	ctx := context.Background()

	// Simple query
	prompt := "What is the capital of France?"
	response, err := llms.GenerateFromSinglePrompt(ctx, llm, prompt)
	if err != nil {
		log.Fatalf("Failed to generate completion: %v", err)
	}

	fmt.Printf("Prompt: %s\n\nResponse: %s\n", prompt, response)

	// Chat completion example
	messages := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeSystem,
			Parts: []llms.ContentPart{
				llms.TextContent{Text: "You are a helpful assistant."},
			},
		},
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextContent{Text: "Tell me about the solar system."},
			},
		},
	}

	chatResponse, err := llm.GenerateContent(ctx, messages)
	if err != nil {
		log.Fatalf("Failed to generate chat completion: %v", err)
	}

	if len(chatResponse.Choices) > 0 {
		fmt.Printf("\nChat Response: %s\n", chatResponse.Choices[0].Content)
	}
}
