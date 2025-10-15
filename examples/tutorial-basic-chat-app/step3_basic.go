package main

import (
	"context"
	"fmt"
	"log"

	"github.com/vendasta/langchaingo/llms"
	"github.com/vendasta/langchaingo/llms/openai"
)

// Step 3: Basic Chat Application
func basicChat() {
	// Initialize the OpenAI LLM
	llm, err := openai.New()
	if err != nil {
		log.Fatal(err)
	}

	// Create a context
	ctx := context.Background()

	// Send a message to the LLM
	response, err := llms.GenerateFromSinglePrompt(
		ctx,
		llm,
		"Hello! How can you help me today?",
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("AI:", response)
}
