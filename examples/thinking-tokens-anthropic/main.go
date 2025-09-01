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

	// Create Anthropic LLM with Claude 3.7+ for extended thinking
	llm, err := anthropic.New(
		anthropic.WithModel("claude-3-7-sonnet"), // Extended thinking model
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("=== Anthropic Thinking Tokens Demo ===\n")

	// Example 1: Basic extended thinking with default settings
	fmt.Println("Example 1: Default Extended Thinking")
	fmt.Println("-------------------------------------")
	
	messages := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextPart("Analyze this logic puzzle: Three boxes are labeled 'Apples', 'Oranges', and 'Apples and Oranges'. Each label is wrong. You can pick one fruit from one box. How do you figure out what's in each box?"),
			},
		},
	}

	resp, err := llm.GenerateContent(ctx, messages,
		llms.WithMaxTokens(500),
		llms.WithThinkingMode(llms.ThinkingModeAuto),
	)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	fmt.Printf("Response: %s\n", resp.Choices[0].Content)
	printTokenUsage(resp.Choices[0].GenerationInfo)

	// Example 2: Low thinking budget for simple task
	fmt.Println("\nExample 2: Low Thinking Budget")
	fmt.Println("-------------------------------")
	
	messages2 := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextPart("What is the capital of France?"),
			},
		},
	}

	resp2, err := llm.GenerateContent(ctx, messages2,
		llms.WithMaxTokens(100),
		llms.WithThinkingBudget(1024), // Minimum budget
	)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	fmt.Printf("Response: %s\n", resp2.Choices[0].Content)
	printTokenUsage(resp2.Choices[0].GenerationInfo)

	// Example 3: High thinking budget for complex reasoning
	fmt.Println("\nExample 3: High Thinking Budget")
	fmt.Println("--------------------------------")
	
	messages3 := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextPart(`Design a distributed system that can:
1. Handle 1 million requests per second
2. Maintain 99.99% uptime
3. Scale automatically based on load
4. Provide global low-latency access

Consider trade-offs between consistency, availability, and partition tolerance.`),
			},
		},
	}

	resp3, err := llm.GenerateContent(ctx, messages3,
		llms.WithMaxTokens(1500),
		llms.WithThinkingBudget(8192), // Large budget for complex reasoning
	)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	fmt.Printf("Response: %s\n", resp3.Choices[0].Content)
	printTokenUsage(resp3.Choices[0].GenerationInfo)

	// Example 4: Return thinking content (if supported)
	fmt.Println("\nExample 4: Return Thinking Content")
	fmt.Println("-----------------------------------")
	
	messages4 := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextPart("Explain step by step how you would debug a memory leak in a Go application."),
			},
		},
	}

	config := &llms.ThinkingConfig{
		Mode:           llms.ThinkingModeMedium,
		BudgetTokens:   4096,
		ReturnThinking: true, // Request thinking content be returned
	}

	resp4, err := llm.GenerateContent(ctx, messages4,
		llms.WithMaxTokens(800),
		llms.WithThinking(config),
	)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	if resp4.Choices[0].ReasoningContent != "" {
		fmt.Println("Thinking Process:")
		fmt.Println("-----------------")
		fmt.Printf("%s\n\n", resp4.Choices[0].ReasoningContent)
	}
	
	fmt.Println("Final Response:")
	fmt.Println("---------------")
	fmt.Printf("%s\n", resp4.Choices[0].Content)
	printTokenUsage(resp4.Choices[0].GenerationInfo)

	// Example 5: Interleaved thinking (Claude 4+ feature)
	fmt.Println("\nExample 5: Interleaved Thinking")
	fmt.Println("--------------------------------")
	
	// Check if using Claude 4+
	modelName := "claude-4" // or "claude-opus-4"
	if llms.IsReasoningModel(modelName) {
		llm4, err := anthropic.New(
			anthropic.WithModel(modelName),
		)
		if err != nil {
			log.Printf("Claude 4 not available: %v", err)
		} else {
			messages5 := []llms.MessageContent{
				{
					Role: llms.ChatMessageTypeHuman,
					Parts: []llms.ContentPart{
						llms.TextPart("Write a haiku about recursion in programming."),
					},
				},
			}

			interleavedConfig := &llms.ThinkingConfig{
				Mode:               llms.ThinkingModeLow,
				InterleaveThinking: true, // Enable interleaved thinking
			}

			resp5, err := llm4.GenerateContent(ctx, messages5,
				llms.WithMaxTokens(200),
				llms.WithThinking(interleavedConfig),
			)
			if err != nil {
				log.Printf("Error: %v", err)
			} else {
				fmt.Printf("Response: %s\n", resp5.Choices[0].Content)
				printTokenUsage(resp5.Choices[0].GenerationInfo)
			}
		}
	} else {
		fmt.Println("Interleaved thinking requires Claude 4+")
	}

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
	
	// Extract thinking token usage
	thinkingUsage := llms.ExtractThinkingTokens(generationInfo)
	if thinkingUsage != nil && thinkingUsage.ThinkingTokens > 0 {
		fmt.Println("\nThinking Token Details:")
		fmt.Printf("  Thinking Tokens: %d\n", thinkingUsage.ThinkingTokens)
		if thinkingUsage.ThinkingInputTokens > 0 {
			fmt.Printf("  Thinking Input: %d\n", thinkingUsage.ThinkingInputTokens)
		}
		if thinkingUsage.ThinkingOutputTokens > 0 {
			fmt.Printf("  Thinking Output: %d\n", thinkingUsage.ThinkingOutputTokens)
		}
		if thinkingUsage.ThinkingBudgetAllocated > 0 {
			fmt.Printf("  Budget Allocated: %d\n", thinkingUsage.ThinkingBudgetAllocated)
			fmt.Printf("  Budget Used: %d\n", thinkingUsage.ThinkingBudgetUsed)
			efficiency := float64(thinkingUsage.ThinkingBudgetUsed) / float64(thinkingUsage.ThinkingBudgetAllocated) * 100
			fmt.Printf("  Efficiency: %.1f%%\n", efficiency)
		}
	}
	
	fmt.Println()
}