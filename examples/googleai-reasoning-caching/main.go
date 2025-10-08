package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/google/generative-ai-go/genai"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/googleai"
)

func main() {
	apiKey := os.Getenv("GOOGLE_API_KEY")
	if apiKey == "" {
		log.Fatal("Please set GOOGLE_API_KEY environment variable")
	}

	ctx := context.Background()

	fmt.Println("=== Google AI Reasoning & Caching Demo ===")
	fmt.Println()

	// Part 1: Demonstrate reasoning support
	demonstrateReasoning(ctx, apiKey)

	fmt.Println()
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println()

	// Part 2: Demonstrate caching support
	demonstrateCaching(ctx, apiKey)
}

func demonstrateReasoning(ctx context.Context, apiKey string) {
	fmt.Println("Part 1: Reasoning Support")
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

	// Check if model supports reasoning
	if reasoner, ok := interface{}(client).(llms.ReasoningModel); ok {
		fmt.Printf("Model supports reasoning: %v\n", reasoner.SupportsReasoning())
	}

	// Test with a complex reasoning problem
	messages := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextPart(`A farmer has 17 sheep. All but 9 of them run away. 
				How many sheep does the farmer have left? Think step by step.`),
			},
		},
	}

	fmt.Println("\nSending reasoning query...")
	resp, err := client.GenerateContent(ctx, messages,
		llms.WithMaxTokens(300),
		llms.WithThinkingMode(llms.ThinkingModeMedium), // Note: Google AI may not use this yet
	)
	if err != nil {
		log.Printf("Error generating content: %v", err)
		return
	}

	if len(resp.Choices) > 0 {
		fmt.Println("\nResponse:")
		fmt.Println(resp.Choices[0].Content)

		// Show token usage
		if genInfo := resp.Choices[0].GenerationInfo; genInfo != nil {
			if tokens, ok := genInfo["TotalTokens"].(int32); ok {
				fmt.Printf("\nTotal tokens used: %d\n", tokens)
			}
		}
	}
}

func demonstrateCaching(ctx context.Context, apiKey string) {
	fmt.Println("Part 2: Caching Support")
	fmt.Println(strings.Repeat("-", 40))
	fmt.Println("Note: Google AI caching requires pre-creating cached content")
	fmt.Println()

	helper := setupCachingHelper(ctx, apiKey)
	cached := createCachedContent(ctx, helper)
	runCachedRequests(ctx, apiKey, cached.Name)
}

func setupCachingHelper(ctx context.Context, apiKey string) *googleai.CachingHelper {
	helper, err := googleai.NewCachingHelper(ctx, googleai.WithAPIKey(apiKey))
	if err != nil {
		log.Fatalf("Failed to create caching helper: %v", err)
	}
	return helper
}

func createCachedContent(ctx context.Context, helper *googleai.CachingHelper) *genai.CachedContent {
	// Create a large context to cache
	largeContext := `You are an expert Go programming assistant.
	You have deep knowledge of Go best practices, performance optimization, and idiomatic code patterns.
	Always consider:
	- Error handling patterns
	- Goroutine safety and concurrency
	- Memory efficiency
	- Code readability and maintainability
	` + strings.Repeat("Always write clean, efficient, and well-documented code. ", 50)

	fmt.Println("Creating cached content...")
	cached, err := helper.CreateCachedContent(ctx, "gemini-2.0-flash",
		[]llms.MessageContent{
			{
				Role: llms.ChatMessageTypeSystem,
				Parts: []llms.ContentPart{
					llms.TextPart(largeContext),
				},
			},
		},
		5*time.Minute, // Cache for 5 minutes
	)
	if err != nil {
		log.Printf("Failed to create cached content: %v", err)
		return nil
	}
	defer func() {
		if err := helper.DeleteCachedContent(ctx, cached.Name); err != nil {
			log.Printf("Failed to delete cached content: %v", err)
		}
	}()

	fmt.Printf("Cached content created: %s\n", cached.Name)
	fmt.Printf("Token count: %d\n", cached.UsageMetadata.TotalTokenCount)
	return cached
}

func runCachedRequests(ctx context.Context, apiKey, cachedContentName string) {
	client, err := googleai.New(ctx, googleai.WithAPIKey(apiKey))
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	questions := []string{
		"What's the best way to handle errors in Go?",
		"How do I avoid goroutine leaks?",
	}

	for i, question := range questions {
		fmt.Printf("\n--- Request %d: %s ---\n", i+1, question)

		messages := []llms.MessageContent{
			{
				Role: llms.ChatMessageTypeHuman,
				Parts: []llms.ContentPart{
					llms.TextPart(question),
				},
			},
		}

		resp, err := client.GenerateContent(ctx, messages,
			googleai.WithCachedContent(cachedContentName),
			llms.WithMaxTokens(200),
		)
		if err != nil {
			log.Printf("Error: %v", err)
			continue
		}

		if len(resp.Choices) > 0 {
			// Show truncated response
			content := resp.Choices[0].Content
			if len(content) > 150 {
				content = content[:150] + "..."
			}
			fmt.Printf("Response: %s\n", content)

			// Check for cached tokens
			if genInfo := resp.Choices[0].GenerationInfo; genInfo != nil {
				if cachedTokens, ok := genInfo["CachedTokens"].(int32); ok && cachedTokens > 0 {
					fmt.Printf("Cached tokens used: %d ✓\n", cachedTokens)
				}
			}
		}
	}

	fmt.Println("\n✨ Google AI caching helps reduce costs by reusing pre-processed context!")
}
