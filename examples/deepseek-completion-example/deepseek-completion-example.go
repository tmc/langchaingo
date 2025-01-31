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
		// If you have a custom API endpoint for Deepseek, you can set it like this:
		// openai.WithBaseURL("https://your-deepseek-endpoint"),
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
		llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
			fmt.Print(string(chunk))
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

		// Print token usage information
		fmt.Printf("\nToken Usage:\n")
		fmt.Printf("- Prompt Tokens: %d\n", completion.Usage.PromptTokens)
		fmt.Printf("- Completion Tokens: %d\n", completion.Usage.CompletionTokens)
		fmt.Printf("- Reasoning Tokens: %d\n", completion.Usage.CompletionTokensDetails.ReasoningTokens)
		fmt.Printf("- Total Tokens: %d\n", completion.Usage.TotalTokens)
	}
}
