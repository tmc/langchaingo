package main

import (
	"context"
	"fmt"
	"log"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)

func main() {
	// Initialize the OpenAI client
	llm, err := openai.New()
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// Create a web search tool with medium context size
	webSearchTool := openai.NewWebSearchTool(openai.WebSearchContextSizeMedium)

	// Create messages asking for current information
	content := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, "You are a helpful assistant with access to web search. Use web search to get current information when needed."),
		llms.TextParts(llms.ChatMessageTypeHuman, "What are the latest developments in AI announced this week?"),
	}

	// Generate content with web search enabled
	completion, err := llm.GenerateContent(
		ctx,
		content,
		llms.WithTools([]llms.Tool{webSearchTool}),
		llms.WithMaxTokens(1000),
	)
	if err != nil {
		log.Fatal(err)
	}

	// Print the response
	if len(completion.Choices) > 0 {
		fmt.Printf("Response:\n%s\n", completion.Choices[0].Content)
	}
}
