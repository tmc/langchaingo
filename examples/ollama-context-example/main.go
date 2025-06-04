package main

import (
	"context"
	"fmt"
	"log"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/ollama"
)

func main() {
	ctx := context.Background()

	// Example 1: Basic usage without context
	fmt.Println("=== Example 1: Basic Usage ===")
	llm, err := ollama.New(ollama.WithModel("llama2"))
	if err != nil {
		log.Fatal(err)
	}

	response, err := llm.GenerateContent(ctx, []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextPart("What is Go programming language?")},
		},
	})
	
	if err != nil {
		log.Printf("Error (this is expected if Ollama is not running): %v", err)
		fmt.Println("Note: Make sure Ollama is running and llama2 model is available")
	} else {
		fmt.Printf("Response: %s\n\n", response.Choices[0].Content)
	}

	// Example 2: Demonstrating context API usage
	fmt.Println("=== Example 2: Context API Demonstration ===")
	
	// Step 1: Create client with context option
	// In a real scenario, this context would come from a previous response
	exampleContext := []int{1, 2, 3, 4, 5} // This would be returned by Ollama
	
	llmWithContext, err := ollama.New(
		ollama.WithModel("llama2"),
		ollama.WithContext(exampleContext),
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Created Ollama client with context: %v\n", exampleContext)
	fmt.Println("This demonstrates the API for using context in future requests.")
	
	// Note: Full functionality requires integration with Ollama's generate API
	fmt.Println("\n=== Context Usage Pattern ===")
	fmt.Println("1. Make initial request: 'Tell me a joke'")
	fmt.Println("2. Extract context from response: [1, 2, 3, ...]")
	fmt.Println("3. Use context in follow-up: 'Tell me another one'")
	fmt.Println("4. Model understands 'another one' refers to another joke")

	// Example 3: Error handling and best practices
	fmt.Println("\n=== Example 3: Best Practices ===")
	
	// Context should be used with the same model
	fmt.Println("✅ Use context with the same model")
	fmt.Println("✅ Check context length limits")
	fmt.Println("✅ Handle context expiration gracefully")
	
	// Demonstrate error case
	_, err = ollama.New(
		ollama.WithModel("llama2"),
		ollama.WithContext(nil), // Empty context is valid
	)
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		fmt.Println("✅ Empty context handled gracefully")
	}

	fmt.Println("\n=== Implementation Status ===")
	fmt.Println("✅ WithContext() option - IMPLEMENTED")
	fmt.Println("⏳ Context extraction from responses - TODO")
	fmt.Println("⏳ Generate API integration - TODO")
	fmt.Println("⏳ Context persistence helpers - TODO")
}

// Example helper function showing how context might be extracted
// This is not yet implemented but shows the intended API
func extractContextFromResponse(response *llms.ContentResponse) []int {
	// TODO: This would extract context from Ollama's response
	// For now, return a placeholder
	return []int{1, 2, 3, 4, 5}
}

// Example helper function for context management
func createContextualClient(model string, previousContext []int) (*ollama.LLM, error) {
	options := []ollama.Option{
		ollama.WithModel(model),
	}
	
	if len(previousContext) > 0 {
		options = append(options, ollama.WithContext(previousContext))
	}
	
	return ollama.New(options...)
}