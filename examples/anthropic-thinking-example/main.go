// Package main demonstrates Anthropic's extended thinking capabilities with Claude 4.
//
// This example shows:
// - Enabling thinking mode for complex reasoning
// - Extracting thinking content and token usage
// - Using different thinking modes (low/medium/high)
// - Interleaved thinking with tool calls
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/anthropic"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	ctx := context.Background()

	// Check for API key
	if os.Getenv("ANTHROPIC_API_KEY") == "" {
		return fmt.Errorf("ANTHROPIC_API_KEY environment variable not set")
	}

	// Create Anthropic client with Claude 4 Sonnet (supports thinking)
	llm, err := anthropic.New(
		anthropic.WithModel("claude-sonnet-4-20250514"),
	)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	// Verify the model supports reasoning
	if !llm.SupportsReasoning() {
		return fmt.Errorf("model does not support reasoning")
	}

	fmt.Println("=== Anthropic Extended Thinking Example ===")

	// Example 1: Basic thinking with medium mode
	if err := exampleBasicThinking(ctx, llm); err != nil {
		return err
	}

	// Example 2: Custom thinking budget
	if err := exampleCustomBudget(ctx, llm); err != nil {
		return err
	}

	// Example 3: Different thinking modes
	if err := exampleThinkingModes(ctx, llm); err != nil {
		return err
	}

	// Example 4: Interleaved thinking with tools
	if err := exampleInterleavedThinking(ctx, llm); err != nil {
		return err
	}

	return nil
}

func exampleBasicThinking(ctx context.Context, llm llms.Model) error {
	fmt.Println("--- Example 1: Basic Thinking (Medium Mode) ---")

	messages := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeHuman,
			`Solve this problem step by step:

A farmer has 17 sheep, and all but 9 die. How many sheep are left?

Think through this carefully and show your reasoning.`),
	}

	resp, err := llm.GenerateContent(ctx, messages,
		llms.WithThinkingMode(llms.ThinkingModeMedium),
		llms.WithTemperature(1.0), // Required for thinking
		llms.WithMaxTokens(4000),
	)
	if err != nil {
		return fmt.Errorf("generate content: %w", err)
	}

	// Extract the response
	choice := resp.Choices[0]
	fmt.Printf("Answer: %s\n\n", choice.Content)

	// Extract thinking usage
	usage := llms.ExtractReasoningUsage(choice.GenerationInfo)
	if usage != nil && usage.ReasoningTokens > 0 {
		fmt.Printf("Thinking Analysis:\n")
		fmt.Printf("  Reasoning Tokens: %d\n", usage.ReasoningTokens)
		fmt.Printf("  Budget Allocated: %d\n", usage.BudgetAllocated)
		fmt.Printf("  Budget Used: %d\n", usage.BudgetUsed)
		fmt.Printf("  Budget Remaining: %d\n\n", usage.BudgetAllocated-usage.BudgetUsed)

		if usage.ThinkingContent != "" {
			fmt.Printf("Thinking Process:\n%s\n\n", usage.ThinkingContent)
		}
	}

	fmt.Println("---")
	return nil
}

func exampleCustomBudget(ctx context.Context, llm llms.Model) error {
	fmt.Println("--- Example 2: Custom Thinking Budget ---")

	messages := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeHuman,
			`What is the most efficient sorting algorithm for nearly-sorted data?
Explain your reasoning process.`),
	}

	// Use custom budget of 5000 tokens
	customBudget := 5000
	resp, err := llm.GenerateContent(ctx, messages,
		llms.WithReasoning(llms.ReasoningOptions{
			BudgetTokens: &customBudget,
		}),
		llms.WithTemperature(1.0),
		llms.WithMaxTokens(4000),
	)
	if err != nil {
		return fmt.Errorf("generate content: %w", err)
	}

	choice := resp.Choices[0]
	fmt.Printf("Answer: %s\n\n", choice.Content)

	usage := llms.ExtractReasoningUsage(choice.GenerationInfo)
	if usage != nil {
		fmt.Printf("Thinking Budget:\n")
		fmt.Printf("  Allocated: %d tokens\n", usage.BudgetAllocated)
		fmt.Printf("  Used: %d tokens\n", usage.BudgetUsed)
		fmt.Printf("  Efficiency: %.1f%%\n\n", float64(usage.BudgetUsed)/float64(usage.BudgetAllocated)*100)
	}

	fmt.Println("---")
	return nil
}

func exampleThinkingModes(ctx context.Context, llm llms.Model) error {
	fmt.Println("--- Example 3: Different Thinking Modes ---")

	problem := "Calculate the 10th Fibonacci number without using recursion."

	modes := []struct {
		mode llms.ThinkingMode
		name string
	}{
		{llms.ThinkingModeLow, "Low"},
		{llms.ThinkingModeMedium, "Medium"},
		{llms.ThinkingModeHigh, "High"},
	}

	for _, m := range modes {
		fmt.Printf("Testing %s Thinking Mode:\n", m.name)

		messages := []llms.MessageContent{
			llms.TextParts(llms.ChatMessageTypeHuman, problem),
		}

		resp, err := llm.GenerateContent(ctx, messages,
			llms.WithThinkingMode(m.mode),
			llms.WithTemperature(1.0),
			llms.WithMaxTokens(2000),
		)
		if err != nil {
			return fmt.Errorf("generate content with %s mode: %w", m.name, err)
		}

		usage := llms.ExtractReasoningUsage(resp.Choices[0].GenerationInfo)
		if usage != nil {
			fmt.Printf("  Reasoning Tokens: %d\n", usage.ReasoningTokens)
			fmt.Printf("  Budget: %d\n\n", usage.BudgetAllocated)
		}
	}

	fmt.Println("---")
	return nil
}

func exampleInterleavedThinking(ctx context.Context, llm llms.Model) error {
	fmt.Println("--- Example 4: Interleaved Thinking with Tools ---")

	// Define tools for the model to use
	tools := []llms.Tool{
		{
			Type: "function",
			Function: &llms.FunctionDefinition{
				Name:        "calculate",
				Description: "Perform a mathematical calculation",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"expression": map[string]any{
							"type":        "string",
							"description": "Mathematical expression to evaluate (e.g., '2 + 2')",
						},
					},
					"required": []string{"expression"},
				},
			},
		},
		{
			Type: "function",
			Function: &llms.FunctionDefinition{
				Name:        "get_data",
				Description: "Retrieve a data point",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"key": map[string]any{
							"type":        "string",
							"description": "Data key to retrieve",
						},
					},
					"required": []string{"key"},
				},
			},
		},
	}

	messages := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeHuman,
			`I need to calculate the total cost of a project.
First get the hourly_rate data, then get the hours_worked data,
and finally calculate the total cost. Show your reasoning at each step.`),
	}

	// Enable interleaved thinking - model will think between tool calls
	resp, err := llm.GenerateContent(ctx, messages,
		llms.WithReasoning(llms.ReasoningOptions{
			Mode:        llms.ThinkingModeMedium,
			Interleaved: true, // Think between tool calls
		}),
		llms.WithTemperature(1.0),
		llms.WithMaxTokens(4000),
		llms.WithTools(tools),
	)
	if err != nil {
		return fmt.Errorf("generate content: %w", err)
	}

	// Check if model requested tools
	if len(resp.Choices[0].ToolCalls) > 0 {
		fmt.Printf("Model requested %d tool calls:\n", len(resp.Choices[0].ToolCalls))
		for i, tc := range resp.Choices[0].ToolCalls {
			var args map[string]any
			json.Unmarshal([]byte(tc.FunctionCall.Arguments), &args)
			fmt.Printf("  %d. %s(%v)\n", i+1, tc.FunctionCall.Name, args)
		}
		fmt.Println()

		// In a real application, you would execute these tools and continue the conversation
		fmt.Println("Note: With interleaved thinking, Claude reasons about which tools to use,")
		fmt.Println("their order, and how to interpret results. This uses additional thinking tokens")
		fmt.Println("but produces more reliable multi-step workflows.")
	}

	usage := llms.ExtractReasoningUsage(resp.Choices[0].GenerationInfo)
	if usage != nil {
		fmt.Printf("Thinking Usage:\n")
		fmt.Printf("  Reasoning Tokens: %d\n", usage.ReasoningTokens)
		if usage.ThinkingContent != "" {
			fmt.Printf("  Thinking Preview: %s...\n", truncate(usage.ThinkingContent, 80))
		}
	}

	fmt.Println("---")
	return nil
}

// Helper function to truncate strings
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen]
}
