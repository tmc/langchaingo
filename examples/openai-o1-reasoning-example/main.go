package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)

var (
	flagModel    = flag.String("model", "o1", "model to use (e.g. 'o1', 'o1-mini', 'o1-preview', 'o3', 'o3-mini')")
	flagStrength = flag.Float64("strength", 0.8, "reasoning strength (0.0=low, 0.5=medium, 1.0=high)")
)

func main() {
	flag.Parse()

	// Create OpenAI client with o1/o3 model
	llm, err := openai.New(
		openai.WithModel(*flagModel),
	)
	if err != nil {
		log.Fatal(err)
	}

	// Check if model supports reasoning
	if !llm.SupportsReasoning() {
		log.Printf("Warning: Model %s may not support reasoning tokens\n", *flagModel)
	}

	ctx := context.Background()

	// Example 1: Complex math problem requiring reasoning
	fmt.Println("=== Example 1: Complex Math Problem ===")
	mathProblem := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeHuman, `
A farmer has 100 feet of fencing and wants to enclose a rectangular area next to a river.
The river will act as one side of the rectangle, so fencing is only needed for the other three sides.
What dimensions will maximize the enclosed area? Show your reasoning step by step.
`),
	}

	resp, err := llm.GenerateContent(ctx, mathProblem,
		llms.WithReasoningStrength(*flagStrength),
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("\n--- Answer ---")
	fmt.Println(resp.Choices[0].Content)

	// Extract reasoning usage
	usage := llms.ExtractReasoningUsage(resp.Choices[0].GenerationInfo)
	if usage != nil {
		fmt.Println("\n--- Reasoning Usage ---")
		fmt.Printf("Reasoning tokens: %d\n", usage.ReasoningTokens)
		fmt.Printf("Output tokens: %d\n", usage.OutputTokens)
		if usage.ReasoningContent != "" {
			fmt.Printf("Reasoning content length: %d characters\n", len(usage.ReasoningContent))
			// Note: For o1/o3 models, reasoning content may not be available
			// as OpenAI doesn't expose the full reasoning process
		}
	} else {
		fmt.Println("\n--- No reasoning usage data available ---")
	}

	// Example 2: Logic puzzle
	fmt.Println("\n\n=== Example 2: Logic Puzzle ===")
	logicPuzzle := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeHuman, `
Three logicians walk into a bar. The bartender asks, "Do all of you want a drink?"
The first logician says, "I don't know."
The second logician says, "I don't know."
The third logician says, "Yes!"

Explain the reasoning behind each answer. What did the bartender's question mean,
and what information did each logician gain from the previous answers?
`),
	}

	resp2, err := llm.GenerateContent(ctx, logicPuzzle,
		llms.WithReasoningStrength(*flagStrength),
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("\n--- Answer ---")
	fmt.Println(resp2.Choices[0].Content)

	// Extract reasoning usage
	usage2 := llms.ExtractReasoningUsage(resp2.Choices[0].GenerationInfo)
	if usage2 != nil {
		fmt.Println("\n--- Reasoning Usage ---")
		fmt.Printf("Reasoning tokens: %d\n", usage2.ReasoningTokens)
		if usage2.ReasoningContent != "" {
			fmt.Printf("Reasoning content: %.100s...\n", usage2.ReasoningContent)
		}
	}

	// Example 3: Different reasoning strengths
	fmt.Println("\n\n=== Example 3: Comparing Reasoning Strengths ===")
	simpleProblem := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeHuman, "What is 15% of 240?"),
	}

	strengths := []float64{0.0, 0.5, 1.0}
	strengthNames := []string{"low", "medium", "high"}

	for i, strength := range strengths {
		fmt.Printf("\n--- With %s reasoning (strength=%.1f) ---\n", strengthNames[i], strength)

		resp3, err := llm.GenerateContent(ctx, simpleProblem,
			llms.WithReasoningStrength(strength),
		)
		if err != nil {
			log.Printf("Error with strength %.1f: %v\n", strength, err)
			continue
		}

		fmt.Println(resp3.Choices[0].Content)

		usage3 := llms.ExtractReasoningUsage(resp3.Choices[0].GenerationInfo)
		if usage3 != nil {
			fmt.Printf("Reasoning tokens used: %d\n", usage3.ReasoningTokens)
		}
	}

	// Show raw generation info for debugging
	fmt.Println("\n\n=== Raw Generation Info (for debugging) ===")
	for key, value := range resp.Choices[0].GenerationInfo {
		fmt.Printf("%s: %v\n", key, value)
	}
}
