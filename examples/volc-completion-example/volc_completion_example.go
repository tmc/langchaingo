package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)

func main() {
	// Initialize OpenAI client configured to use Volc Cloud API
	llm, err := openai.New(
		// Set Volc Cloud API base URL
		openai.WithBaseURL("https://ark.cn-beijing.volces.com/api/v3"),
		// Set API key
		openai.WithToken(""),
		// Set model endpoint ID
		openai.WithModel(""),
	)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// Create chat messages
	content := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, "You are a helpful assistant that explains complex topics step by step"),
		llms.TextParts(llms.ChatMessageTypeHuman, "Explain how quantum entanglement works and why it's important for quantum computing"),
	}

	// Variable to store content chunks
	var contentBuilder strings.Builder

	fmt.Println("Generating response, please wait...")
	fmt.Println("----------------------------------------")
	fmt.Println("\nReasoning process:")

	// Generate content with streaming to see reasoning process and final answer in real-time
	completion, err := llm.GenerateContent(
		ctx,
		content,
		llms.WithMaxTokens(2000),
		llms.WithTemperature(0.7),
		llms.WithStreamingReasoningFunc(func(ctx context.Context, reasoningChunk []byte, chunk []byte) error {
			if len(reasoningChunk) > 0 {
				// Print reasoning chunks directly as they arrive
				fmt.Print(string(reasoningChunk))
			}
			if len(chunk) > 0 {
				// Store content chunks for later display
				contentBuilder.Write(chunk)
			}
			return nil
		}),
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("\n----------------------------------------")

	// Access the final answer
	if len(completion.Choices) > 0 {
		choice := completion.Choices[0]

		// Output the final answer
		fmt.Println("\n[FINAL ANSWER]:")
		fmt.Println("----------------------------------------")
		if contentBuilder.Len() > 0 {
			fmt.Println(contentBuilder.String())
		} else {
			fmt.Println(choice.Content)
		}
	}
}
