package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/vendasta/langchaingo/llms"
	"github.com/vendasta/langchaingo/llms/anthropic"
	"github.com/vendasta/langchaingo/llms/openai"
)

func main() {
	ctx := context.Background()

	fmt.Println("=== Multi-Backend Reasoning/Thinking Tokens Demo ===")
	fmt.Println("Comparing reasoning capabilities across LLM providers")
	fmt.Println()

	// Complex reasoning prompt that benefits from step-by-step thinking
	prompt := `A farmer needs to cross a river with a fox, a chicken, and a bag of grain. 
The boat can only carry the farmer and one item at a time. 
If left alone, the fox will eat the chicken, and the chicken will eat the grain. 
How can the farmer get everything across safely? Think through this step-by-step.`

	// Test with OpenAI o1-mini (reasoning model)
	if apiKey := os.Getenv("OPENAI_API_KEY"); apiKey != "" {
		fmt.Println(strings.Repeat("=", 60))
		fmt.Println("OpenAI o1-mini (Reasoning Model)")
		fmt.Println(strings.Repeat("=", 60))

		llm, err := openai.New(openai.WithModel("o1-mini"))
		if err != nil {
			fmt.Printf("Error initializing OpenAI: %v\n\n", err)
		} else {
			testReasoning(ctx, llm, "o1-mini", prompt, true)
		}
	}

	// Test with Anthropic Claude 3.7 (supports extended thinking)
	if apiKey := os.Getenv("ANTHROPIC_API_KEY"); apiKey != "" {
		fmt.Println(strings.Repeat("=", 60))
		fmt.Println("Anthropic Claude 3.7 Sonnet (Extended Thinking)")
		fmt.Println(strings.Repeat("=", 60))

		llm, err := anthropic.New(anthropic.WithModel("claude-3-7-sonnet-20250219"))
		if err != nil {
			fmt.Printf("Error initializing Anthropic: %v\n\n", err)
		} else {
			testReasoning(ctx, llm, "claude-3-7-sonnet-20250219", prompt, true)
		}
	}

	// Compare with standard models (no reasoning tokens)
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("COMPARISON: Standard Models (No Reasoning)")
	fmt.Println(strings.Repeat("=", 60))

	if apiKey := os.Getenv("OPENAI_API_KEY"); apiKey != "" {
		fmt.Println("\n--- OpenAI GPT-4 Turbo ---")
		llm, err := openai.New(openai.WithModel("gpt-4-turbo-preview"))
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		} else {
			testReasoning(ctx, llm, "gpt-4-turbo-preview", prompt, false)
		}
	}

	if apiKey := os.Getenv("ANTHROPIC_API_KEY"); apiKey != "" {
		fmt.Println("\n--- Anthropic Claude 3 Sonnet ---")
		llm, err := anthropic.New(anthropic.WithModel("claude-3-sonnet-20240229"))
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		} else {
			testReasoning(ctx, llm, "claude-3-sonnet-20240229", prompt, false)
		}
	}

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("Demo Complete!")
	fmt.Println("\nKey Observations:")
	fmt.Println("- Reasoning models use additional 'thinking' tokens for internal processing")
	fmt.Println("- These tokens improve response quality but aren't shown in the output")
	fmt.Println("- Standard models generate all tokens as visible output")
}

func testReasoning(ctx context.Context, llm llms.Model, modelName string, prompt string, expectReasoning bool) {
	// Check if model supports reasoning
	supportsReasoning := false
	if reasoner, ok := llm.(llms.ReasoningModel); ok {
		supportsReasoning = reasoner.SupportsReasoning()
	}

	fmt.Printf("Model: %s\n", modelName)
	fmt.Printf("Reasoning Support: %v\n", supportsReasoning)

	if expectReasoning && !supportsReasoning {
		fmt.Println("Note: This model version may not support reasoning tokens yet")
	}

	messages := []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextPart(prompt)},
		},
	}

	// Configure thinking mode based on model capabilities
	opts := []llms.CallOption{
		llms.WithMaxTokens(500),
	}

	if supportsReasoning {
		// Use high thinking mode for complex reasoning
		opts = append(opts, llms.WithThinkingMode(llms.ThinkingModeHigh))
		fmt.Println("Thinking Mode: HIGH (maximum reasoning depth)")
	} else {
		fmt.Println("Thinking Mode: NONE (standard generation)")
	}

	fmt.Print("\nGenerating response... ")
	resp, err := llm.GenerateContent(ctx, messages, opts...)
	if err != nil {
		fmt.Printf("Error: %v\n\n", err)
		return
	}
	fmt.Println("done")

	// Display response
	content := resp.Choices[0].Content
	fmt.Println("\n--- Response ---")
	if len(content) > 400 {
		fmt.Println(content[:400] + "...\n[truncated]")
	} else {
		fmt.Println(content)
	}

	// Display detailed token metrics
	fmt.Println("\n--- Token Metrics ---")
	if genInfo := resp.Choices[0].GenerationInfo; genInfo != nil {
		// Extract standard token usage
		var inputTokens, outputTokens, totalTokens int
		if v, ok := genInfo["PromptTokens"].(int); ok {
			inputTokens = v
		}
		if v, ok := genInfo["CompletionTokens"].(int); ok {
			outputTokens = v
		}
		if v, ok := genInfo["TotalTokens"].(int); ok {
			totalTokens = v
		}

		fmt.Printf("Input Tokens:      %d\n", inputTokens)
		fmt.Printf("Output Tokens:     %d\n", outputTokens)
		fmt.Printf("Total Tokens:      %d\n", totalTokens)

		// Extract thinking-specific tokens
		usage := llms.ExtractThinkingTokens(genInfo)
		if usage != nil && usage.ThinkingTokens > 0 {
			fmt.Printf("\nReasoning Breakdown:\n")
			fmt.Printf("  Thinking Tokens:   %d\n", usage.ThinkingTokens)
			fmt.Printf("  Visible Output:    %d\n", outputTokens-usage.ThinkingTokens)
			fmt.Printf("  Thinking Ratio:    %.1f%% of output\n",
				float64(usage.ThinkingTokens)/float64(outputTokens)*100)

			if usage.ThinkingBudgetAllocated > 0 {
				fmt.Printf("  Budget Allocated:  %d\n", usage.ThinkingBudgetAllocated)
				fmt.Printf("  Budget Used:       %d\n", usage.ThinkingBudgetUsed)
				fmt.Printf("  Budget Efficiency: %.1f%%\n",
					float64(usage.ThinkingBudgetUsed)/float64(usage.ThinkingBudgetAllocated)*100)
			}
		} else if supportsReasoning {
			fmt.Println("\nNote: Model supports reasoning but no thinking tokens were used.")
			fmt.Println("      This may happen for simpler prompts or certain model versions.")
		}
	}
	fmt.Println()
}
