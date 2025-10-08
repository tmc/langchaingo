package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/vendasta/langchaingo/llms"
	"github.com/vendasta/langchaingo/llms/anthropic"
)

func main() {
	ctx := context.Background()

	fmt.Println("=== Claude 3.7+ Extended Capabilities Demo ===")
	fmt.Println("Demonstrating combined extended thinking + 128K output")
	fmt.Println()

	// Complex prompt that benefits from both extended thinking and long output
	prompt := `Write a comprehensive technical guide about building distributed systems.

Include:
1. Theoretical foundations (CAP theorem, consensus algorithms)
2. Practical implementation patterns
3. Real-world case studies from major tech companies
4. Code examples in Go for key concepts
5. Testing strategies for distributed systems
6. Common pitfalls and how to avoid them
7. Performance optimization techniques
8. Monitoring and observability best practices

Make this guide as detailed and comprehensive as possible, targeting senior engineers
who want to deeply understand distributed systems architecture.`

	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		fmt.Println("ANTHROPIC_API_KEY not set. Skipping demo.")
		fmt.Println("Set the environment variable to run this example:")
		fmt.Println("  export ANTHROPIC_API_KEY=your-api-key")
		return
	}

	// Initialize Claude 3.7 with extended capabilities
	llm, err := anthropic.New(
		anthropic.WithModel("claude-3-7-sonnet-20250219"),
	)
	if err != nil {
		fmt.Printf("Error initializing Anthropic: %v\n", err)
		return
	}

	messages := []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextPart(prompt)},
		},
	}

	// Configure with both extended thinking AND extended output
	opts := []llms.CallOption{
		// Enable extended thinking for complex reasoning
		llms.WithThinkingMode(llms.ThinkingModeHigh),
		// Enable 128K output for comprehensive response
		anthropic.WithExtendedOutput(),
		// Set high token limit to utilize extended output
		llms.WithMaxTokens(50000), // Can go up to 128K
		// Temperature must be 1 when thinking is enabled
		llms.WithTemperature(1.0),
	}

	fmt.Println("Generating comprehensive guide with:")
	fmt.Println("  • Extended thinking (HIGH mode)")
	fmt.Println("  • Extended output (up to 128K tokens)")
	fmt.Println("  • Max tokens set to 50,000")
	fmt.Println()
	fmt.Print("Processing (this may take a while)... ")

	resp, err := llm.GenerateContent(ctx, messages, opts...)
	if err != nil {
		fmt.Printf("\nError: %v\n", err)
		return
	}
	fmt.Println("done")

	// Display response summary
	content := resp.Choices[0].Content
	contentLen := len(content)
	
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("RESPONSE SUMMARY")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("Total response length: %d characters\n", contentLen)
	
	// Show first 1000 chars as preview
	preview := content
	if len(preview) > 1000 {
		preview = preview[:1000] + "..."
	}
	fmt.Println("\nPreview (first 1000 chars):")
	fmt.Println(strings.Repeat("-", 40))
	fmt.Println(preview)
	fmt.Println(strings.Repeat("-", 40))

	// Display detailed token metrics
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("TOKEN METRICS")
	fmt.Println(strings.Repeat("=", 60))
	
	if genInfo := resp.Choices[0].GenerationInfo; genInfo != nil {
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

		fmt.Printf("Input Tokens:       %d\n", inputTokens)
		fmt.Printf("Output Tokens:      %d\n", outputTokens)
		fmt.Printf("Total Tokens:       %d\n", totalTokens)

		// Check for thinking tokens
		usage := llms.ExtractThinkingTokens(genInfo)
		if usage != nil && usage.ThinkingTokens > 0 {
			fmt.Printf("\nThinking Analysis:\n")
			fmt.Printf("  Thinking Tokens:    %d\n", usage.ThinkingTokens)
			fmt.Printf("  Visible Output:     %d\n", outputTokens-usage.ThinkingTokens)
			fmt.Printf("  Thinking Ratio:     %.1f%% of output\n",
				float64(usage.ThinkingTokens)/float64(outputTokens)*100)

			if usage.ThinkingBudgetAllocated > 0 {
				fmt.Printf("  Budget Allocated:   %d\n", usage.ThinkingBudgetAllocated)
				fmt.Printf("  Budget Used:        %d\n", usage.ThinkingBudgetUsed)
			}
		}

		// Highlight extended output usage
		if outputTokens > 8192 {
			fmt.Printf("\n✅ Extended Output Active: Generated %d tokens (standard limit is 8192)\n", outputTokens)
		}
	}

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("Demo complete!")
	fmt.Println("\nKey Features Demonstrated:")
	fmt.Println("• Extended thinking for complex reasoning about distributed systems")
	fmt.Println("• Extended output allowing comprehensive, detailed responses")
	fmt.Println("• Combined capabilities working together seamlessly")
	
	// Optionally save full response to file
	if contentLen > 10000 {
		filename := "distributed-systems-guide.md"
		fmt.Printf("\nFull response is %d chars. Save to %s? (y/n): ", contentLen, filename)
		var answer string
		fmt.Scanln(&answer)
		if strings.ToLower(answer) == "y" {
			err := os.WriteFile(filename, []byte(content), 0644)
			if err != nil {
				fmt.Printf("Error saving file: %v\n", err)
			} else {
				fmt.Printf("Saved to %s\n", filename)
			}
		}
	}
}