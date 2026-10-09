package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/ollama"
)

func main() {
	// Get Ollama server URL from environment or use default
	serverURL := os.Getenv("OLLAMA_HOST")
	if serverURL == "" {
		serverURL = "http://localhost:11434"
	}

	// Create Ollama client with a reasoning-capable model
	// Models like deepseek-r1, qwq, or any model with "thinking" support
	llm, err := ollama.New(
		ollama.WithServerURL(serverURL),
		ollama.WithModel("gpt-oss:latest"),
	)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// Complex reasoning problem
	messages := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextPart(`Alice has 3 apples. Bob has twice as many apples as Alice.
Charlie has 2 fewer apples than Bob. How many apples do they have in total?
Think through this step by step.`),
			},
		},
	}

	fmt.Println("Sending reasoning query...")
	resp, err := llm.GenerateContent(ctx, messages,
		llms.WithMaxTokens(500),
		llms.WithThinkingMode(llms.ThinkingModeMedium), // Enable thinking if supported
	)
	if err != nil {
		log.Fatal(err)
	}

	if len(resp.Choices) == 0 {
		log.Fatal("no response choices returned")
	}

	choice := resp.Choices[0]

	// Display thinking content if available
	if choice.ReasoningContent != "" {
		fmt.Println("\n=== Thinking Process ===")
		fmt.Println(choice.ReasoningContent)
	}

	// Display the final response
	fmt.Println("\n=== Response ===")
	fmt.Println(choice.Content)

	// Display generation info
	if genInfo := choice.GenerationInfo; genInfo != nil {
		fmt.Println("\n=== Generation Info ===")
		if tokens, ok := genInfo["TotalTokens"].(int); ok {
			fmt.Printf("Total tokens: %d\n", tokens)
		}
		if enabled, ok := genInfo["ThinkingEnabled"].(bool); ok && enabled {
			fmt.Println("Thinking mode: enabled")
		}
		if thinking, ok := genInfo["ThinkingContent"].(string); ok && thinking != "" {
			fmt.Printf("Thinking content length: %d chars\n", len(thinking))
		}
	}
}
