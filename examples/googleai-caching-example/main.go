package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/llms/googleai"
)

func main() {
	ctx := context.Background()

	// Get API key from environment
	apiKey := os.Getenv("GOOGLE_API_KEY")
	if apiKey == "" {
		log.Fatal("GOOGLE_API_KEY environment variable not set")
	}

	// Create a caching helper
	helper, err := googleai.NewCachingHelper(ctx, googleai.WithAPIKey(apiKey))
	if err != nil {
		log.Fatal(err)
	}

	// Create a large context to cache (minimum 32,768 tokens required)
	// This example creates a comprehensive guide about Go programming
	baseContext := `You are an expert Go programming assistant with deep knowledge of Go best practices, 
idioms, and the Go ecosystem. You specialize in:

1. Go Language Fundamentals:
   - Understanding of Go's type system, interfaces, and composition
   - Goroutines, channels, and concurrency patterns
   - Memory management and the garbage collector
   - Error handling patterns and best practices

2. Standard Library Expertise:
   - Proficiency with common packages: fmt, io, net/http, encoding/json
   - Understanding of context package for cancellation and deadlines
   - File I/O and system interaction

3. Code Quality:
   - Writing idiomatic Go code following community conventions
   - Effective error handling and propagation
   - Proper use of defer, panic, and recover
   - Testing with the testing package

4. Performance Optimization:
   - Profiling with pprof
   - Benchmark writing and interpretation
   - Common performance pitfalls and solutions
   - Memory allocation optimization

5. Common Patterns:
   - Functional options pattern
   - Worker pools and fan-out/fan-in
   - Circuit breakers and retry logic
   - Dependency injection

6. Tools and Ecosystem:
   - Go modules and dependency management
   - Common linters: golangci-lint, staticcheck
   - Code generation tools
   - Popular frameworks and libraries

When reviewing code or answering questions, always:
- Prioritize clarity and maintainability
- Consider error handling carefully
- Think about concurrent access and race conditions
- Follow established Go conventions
- Provide practical, runnable examples when possible

`

	// Repeat the context to reach the minimum cache size (~32k tokens)
	// Each repetition adds context about different aspects of Go
	longContext := strings.Repeat(baseContext, 500)

	fmt.Println("Creating cached content (this may take a moment)...")
	cached, err := helper.CreateCachedContent(
		ctx,
		"gemini-2.5-pro",
		[]llms.MessageContent{
			{
				Role: llms.ChatMessageTypeSystem,
				Parts: []llms.ContentPart{
					llms.TextPart(longContext),
				},
			},
		},
		1*time.Hour, // Cache for 1 hour
		"go-expert-assistant",
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("✓ Created cached content: %s\n", cached.Name)
	fmt.Printf("  Display Name: %s\n", cached.DisplayName)
	fmt.Printf("  Model: %s\n", cached.Model)
	fmt.Printf("  Total Tokens: %d\n", cached.UsageMetadata.TotalTokenCount)
	fmt.Printf("  Expires: %s\n", cached.ExpireTime.Format(time.RFC3339))
	fmt.Println()

	// Cleanup cached content when done
	defer func() {
		fmt.Println("\nCleaning up cached content...")
		if err := helper.DeleteCachedContent(ctx, cached.Name); err != nil {
			log.Printf("Failed to delete cached content: %v", err)
		} else {
			fmt.Println("✓ Cached content deleted")
		}
	}()

	// Now use the cached content in a request
	fmt.Println("Making request with cached content...")
	client, err := googleai.New(ctx,
		googleai.WithAPIKey(apiKey),
		googleai.WithDefaultModel("gemini-2.5-pro"), // Must match cached content model
	)
	if err != nil {
		log.Fatal(err)
	}

	// Example 1: Code review question
	fmt.Println("\n--- Example 1: Code Review ---")
	resp, err := client.GenerateContent(
		ctx,
		[]llms.MessageContent{
			{
				Role: llms.ChatMessageTypeHuman,
				Parts: []llms.ContentPart{
					llms.TextPart("What should I check when reviewing error handling in Go code?"),
				},
			},
		},
		googleai.WithCachedContent(cached.Name),
		llms.WithMaxTokens(300),
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(resp.Choices[0].Content)
	if genInfo := resp.Choices[0].GenerationInfo; genInfo != nil {
		if cachedTokens, ok := genInfo["PromptCachedTokens"].(int); ok {
			fmt.Printf("\n[Cached tokens used: %d]\n", cachedTokens)
		}
	}

	// Example 2: Performance question (reusing the same cache)
	fmt.Println("\n--- Example 2: Performance Optimization ---")
	resp2, err := client.GenerateContent(
		ctx,
		[]llms.MessageContent{
			{
				Role: llms.ChatMessageTypeHuman,
				Parts: []llms.ContentPart{
					llms.TextPart("How can I reduce memory allocations in my Go application?"),
				},
			},
		},
		googleai.WithCachedContent(cached.Name),
		llms.WithMaxTokens(300),
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(resp2.Choices[0].Content)
	if genInfo := resp2.Choices[0].GenerationInfo; genInfo != nil {
		if cachedTokens, ok := genInfo["PromptCachedTokens"].(int); ok {
			fmt.Printf("\n[Cached tokens used: %d]\n", cachedTokens)
		}
	}

	// List all cached contents
	fmt.Println("\n--- Listing All Cached Contents ---")
	count := 0
	for cache, err := range helper.AllCachedContents(ctx) {
		if err != nil {
			log.Printf("Error listing cache: %v", err)
			continue
		}
		count++
		fmt.Printf("%d. %s (Display: %s, Tokens: %d)\n",
			count,
			cache.Name,
			cache.DisplayName,
			cache.UsageMetadata.TotalTokenCount,
		)
		if count >= 5 {
			fmt.Println("   (showing first 5 only...)")
			break
		}
	}
}
