package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/googleai"
)

func main() {
	apiKey := os.Getenv("GOOGLE_API_KEY")
	if apiKey == "" {
		log.Fatal("Please set GOOGLE_API_KEY environment variable")
	}

	ctx := context.Background()

	fmt.Println("=== Google AI Reasoning Support Demo ===")
	fmt.Println()
	fmt.Println("This example demonstrates reasoning support for Gemini 2.0+ models.")
	fmt.Println("Note: Google's reasoning implementation differs from Anthropic's.")
	fmt.Println("Gemini 2.0+ has inherent reasoning capabilities but doesn't expose")
	fmt.Println("explicit thinking mode controls or separate thinking tokens yet.")
	fmt.Println()
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println()

	// Part 1: Check reasoning capability
	demonstrateCapabilityCheck(ctx, apiKey)

	fmt.Println()
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println()

	// Part 2: Use reasoning with complex problems
	demonstrateReasoning(ctx, apiKey)

	fmt.Println()
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println()

	// Part 3: Extract reasoning usage information
	demonstrateReasoningExtraction(ctx, apiKey)
}

func demonstrateCapabilityCheck(ctx context.Context, apiKey string) {
	fmt.Println("Part 1: Reasoning Capability Check")
	fmt.Println(strings.Repeat("-", 40))

	// Create clients with different models
	models := []struct {
		name      string
		modelName string
	}{
		{"Gemini 2.0 Flash", "gemini-2.0-flash"},
		{"Gemini 1.5 Flash", "gemini-1.5-flash"},
	}

	for _, m := range models {
		client, err := googleai.New(ctx,
			googleai.WithAPIKey(apiKey),
			googleai.WithDefaultModel(m.modelName),
		)
		if err != nil {
			log.Printf("Failed to create client for %s: %v", m.name, err)
			continue
		}
		defer client.Close()

		// Check if model supports reasoning
		if reasoner, ok := interface{}(client).(llms.ReasoningModel); ok {
			supports := reasoner.SupportsReasoning()
			fmt.Printf("%s (%s): Reasoning support = %v\n", m.name, m.modelName, supports)
		}
	}
}

func demonstrateReasoning(ctx context.Context, apiKey string) {
	fmt.Println("Part 2: Using Reasoning with Complex Problems")
	fmt.Println(strings.Repeat("-", 40))

	// Create client with Gemini 2.0 Flash (supports reasoning)
	client, err := googleai.New(ctx,
		googleai.WithAPIKey(apiKey),
		googleai.WithDefaultModel("gemini-2.0-flash"),
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	// Problem 1: Logic puzzle
	fmt.Println("\nProblem 1: Logic Puzzle")
	fmt.Println("Question: A farmer has 17 sheep. All but 9 of them run away.")
	fmt.Println("          How many sheep does the farmer have left?")
	fmt.Println()

	messages1 := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextPart(`A farmer has 17 sheep. All but 9 of them run away.
How many sheep does the farmer have left? Think step by step and explain your reasoning.`),
			},
		},
	}

	resp1, err := client.GenerateContent(ctx, messages1,
		llms.WithMaxTokens(300),
		llms.WithThinkingMode(llms.ThinkingModeMedium),
	)
	if err != nil {
		log.Printf("Error: %v", err)
	} else if len(resp1.Choices) > 0 {
		fmt.Printf("Answer:\n%s\n", resp1.Choices[0].Content)
		printTokenUsage(resp1.Choices[0].GenerationInfo)
	}

	// Problem 2: Mathematical reasoning
	fmt.Println()
	fmt.Println(strings.Repeat("-", 40))
	fmt.Println("\nProblem 2: Mathematical Reasoning")
	fmt.Println("Question: What is 347 * 29? Show your work.")
	fmt.Println()

	messages2 := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextPart(`What is 347 * 29? Break down the calculation step by step.`),
			},
		},
	}

	resp2, err := client.GenerateContent(ctx, messages2,
		llms.WithMaxTokens(400),
		llms.WithThinkingMode(llms.ThinkingModeHigh),
	)
	if err != nil {
		log.Printf("Error: %v", err)
	} else if len(resp2.Choices) > 0 {
		fmt.Printf("Answer:\n%s\n", resp2.Choices[0].Content)
		printTokenUsage(resp2.Choices[0].GenerationInfo)
	}

	// Problem 3: Complex word problem
	fmt.Println()
	fmt.Println(strings.Repeat("-", 40))
	fmt.Println("\nProblem 3: Complex Word Problem")
	fmt.Println("Question: Alice has 3 times as many apples as Bob.")
	fmt.Println("          Bob has 5 fewer apples than Charlie.")
	fmt.Println("          Charlie has 12 apples. How many apples does Alice have?")
	fmt.Println()

	messages3 := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextPart(`Alice has 3 times as many apples as Bob.
Bob has 5 fewer apples than Charlie.
Charlie has 12 apples.
How many apples does Alice have? Explain your reasoning step by step.`),
			},
		},
	}

	resp3, err := client.GenerateContent(ctx, messages3,
		llms.WithMaxTokens(400),
		llms.WithThinkingMode(llms.ThinkingModeMedium),
	)
	if err != nil {
		log.Printf("Error: %v", err)
	} else if len(resp3.Choices) > 0 {
		fmt.Printf("Answer:\n%s\n", resp3.Choices[0].Content)
		printTokenUsage(resp3.Choices[0].GenerationInfo)
	}
}

func demonstrateReasoningExtraction(ctx context.Context, apiKey string) {
	fmt.Println("Part 3: Extracting Reasoning Usage")
	fmt.Println(strings.Repeat("-", 40))
	fmt.Println()
	fmt.Println("Note: Google AI doesn't currently expose separate thinking tokens")
	fmt.Println("like Anthropic Claude 4. The ExtractReasoningUsage() function")
	fmt.Println("returns standardized fields for cross-provider compatibility.")
	fmt.Println()

	client, err := googleai.New(ctx,
		googleai.WithAPIKey(apiKey),
		googleai.WithDefaultModel("gemini-2.0-flash"),
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	messages := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextPart("What is 15 * 23? Think carefully."),
			},
		},
	}

	resp, err := client.GenerateContent(ctx, messages,
		llms.WithMaxTokens(200),
		llms.WithReasoning(llms.ReasoningOptions{
			Mode: llms.ThinkingModeMedium,
		}),
	)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	if len(resp.Choices) > 0 {
		// Extract reasoning usage using the standardized function
		usage := llms.ExtractReasoningUsage(resp.Choices[0].GenerationInfo)

		fmt.Println("Reasoning Usage Information:")
		if usage != nil {
			fmt.Printf("  Reasoning Tokens: %d\n", usage.ReasoningTokens)
			fmt.Printf("  Output Tokens: %d\n", usage.OutputTokens)
			if usage.ReasoningContent != "" {
				fmt.Printf("  Reasoning Content: %s\n", usage.ReasoningContent)
			}
			if usage.ThinkingContent != "" {
				fmt.Printf("  Thinking Content: %s\n", usage.ThinkingContent)
			}
		} else {
			fmt.Println("  (No reasoning usage data available)")
		}

		// Show raw GenerationInfo for comparison
		fmt.Println()
		fmt.Println("Raw Generation Info:")
		if genInfo := resp.Choices[0].GenerationInfo; genInfo != nil {
			fmt.Printf("  Total Tokens: %v\n", genInfo["TotalTokens"])
			fmt.Printf("  Thinking Tokens: %v\n", genInfo["ThinkingTokens"])
			fmt.Printf("  Thinking Content: %q\n", genInfo["ThinkingContent"])
		}
	}
}

func printTokenUsage(genInfo map[string]any) {
	if genInfo == nil {
		return
	}

	fmt.Println()
	fmt.Println("Token Usage:")
	if totalTokens, ok := genInfo["TotalTokens"].(int32); ok {
		fmt.Printf("  Total: %d tokens\n", totalTokens)
	}
	if promptTokens, ok := genInfo["PromptTokens"].(int32); ok {
		fmt.Printf("  Input: %d tokens\n", promptTokens)
	}
	if completionTokens, ok := genInfo["CompletionTokens"].(int32); ok {
		fmt.Printf("  Output: %d tokens\n", completionTokens)
	}

	// Note: Google doesn't expose thinking tokens separately yet
	if thinkingTokens, ok := genInfo["ThinkingTokens"].(int); ok && thinkingTokens > 0 {
		fmt.Printf("  Thinking: %d tokens\n", thinkingTokens)
	}
}
