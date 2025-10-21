package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/ollama"
)

func main() {
	fmt.Println("=== Ollama Reasoning & Caching Demo ===")
	fmt.Println()

	// Get Ollama server URL from environment or use default
	serverURL := os.Getenv("OLLAMA_HOST")
	if serverURL == "" {
		serverURL = "http://localhost:11434"
	}

	ctx := context.Background()

	// Part 1: Demonstrate reasoning support
	demonstrateReasoning(ctx, serverURL)

	fmt.Println()
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println()

	// Part 2: Demonstrate context caching
	demonstrateCaching(ctx, serverURL)
}

func demonstrateReasoning(ctx context.Context, serverURL string) {
	fmt.Println("Part 1: Reasoning Support")
	fmt.Println(strings.Repeat("-", 40))
	fmt.Println("Note: Requires a reasoning model like deepseek-r1 or qwq")
	fmt.Println()

	// Try to use a reasoning model
	modelName := "deepseek-r1:latest"
	fmt.Printf("Attempting to use model: %s\n", modelName)

	llm, err := ollama.New(
		ollama.WithServerURL(serverURL),
		ollama.WithModel(modelName),
		ollama.WithPullModel(),                // Auto-pull if not available
		ollama.WithPullTimeout(5*time.Minute), // Allow time for download
	)
	if err != nil {
		log.Printf("Failed to create Ollama client: %v", err)
		fmt.Println("Falling back to a standard model for demo...")

		// Fall back to a standard model
		modelName = "llama3:latest"
		llm, err = ollama.New(
			ollama.WithServerURL(serverURL),
			ollama.WithModel(modelName),
		)
		if err != nil {
			log.Fatalf("Failed to create fallback client: %v", err)
		}
	}

	// Check if model supports reasoning
	if reasoner, ok := interface{}(llm).(llms.ReasoningModel); ok {
		fmt.Printf("Model supports reasoning: %v\n", reasoner.SupportsReasoning())
	}

	// Complex reasoning problem
	messages := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextPart(`Alice has 3 apples. Bob has twice as many apples as Alice. 
				Charlie has 2 fewer apples than Bob. How many apples do they have in total?
				Think through this step by step.`),
			},
		},
	}

	fmt.Println("\nSending reasoning query...")
	resp, err := llm.GenerateContent(ctx, messages,
		llms.WithMaxTokens(300),
		llms.WithThinkingMode(llms.ThinkingModeMedium), // Enable thinking if supported
	)
	if err != nil {
		log.Printf("Error generating content: %v", err)
		return
	}

	if len(resp.Choices) > 0 {
		fmt.Println("\nResponse:")
		content := resp.Choices[0].Content
		// Truncate very long responses
		if len(content) > 500 {
			content = content[:500] + "...\n[truncated]"
		}
		fmt.Println(content)

		// Show token usage
		if genInfo := resp.Choices[0].GenerationInfo; genInfo != nil {
			if tokens, ok := genInfo["TotalTokens"].(int); ok {
				fmt.Printf("\nTotal tokens used: %d\n", tokens)
			}
			if enabled, ok := genInfo["ThinkingEnabled"].(bool); ok && enabled {
				fmt.Println("✓ Thinking mode was enabled")
			}
		}
	}
}

func demonstrateCaching(ctx context.Context, serverURL string) {
	fmt.Println("Part 2: Context Caching")
	fmt.Println(strings.Repeat("-", 40))
	fmt.Println("Using in-memory context cache to reduce redundant processing")
	fmt.Println()

	llm, cache := setupCachingClient(serverURL)
	runCachingDemo(ctx, llm, cache)
}

func setupCachingClient(serverURL string) (llms.Model, *ollama.ContextCache) {
	// Create Ollama client
	llm, err := ollama.New(
		ollama.WithServerURL(serverURL),
		ollama.WithModel("llama3:latest"), // Use any available model
	)
	if err != nil {
		log.Fatalf("Failed to create Ollama client: %v", err)
	}

	// Create context cache
	cache := ollama.NewContextCache(10, 10*time.Minute)
	fmt.Println("Created context cache (10 entries, 10 min TTL)")

	return llm, cache
}

func runCachingDemo(ctx context.Context, llm llms.Model, cache *ollama.ContextCache) {
	// Large system context that benefits from caching
	systemContext := `You are an expert programming assistant specializing in Go.
	You have deep knowledge of Go idioms, best practices, and performance optimization.
	Always provide clear, concise, and idiomatic Go code examples.`

	// Multiple questions using the same context
	questions := []string{
		"What's the best way to handle errors in Go?",
		"How do I create a goroutine-safe map?",
		"What are channels used for in Go?",
	}

	for i, question := range questions {
		processQuestion(ctx, llm, cache, systemContext, question, i+1)
	}

	// Show cache statistics
	entries, hits, avgSaved := cache.Stats()
	fmt.Printf("\n=== Cache Statistics ===\n")
	fmt.Printf("Entries: %d\n", entries)
	fmt.Printf("Total hits: %d\n", hits)
	fmt.Printf("Avg tokens saved per hit: %d\n", avgSaved)

	if hits > 0 {
		fmt.Println("\n✨ Context caching reduced token processing and improved response times!")
	}
}

func processQuestion(ctx context.Context, llm llms.Model, cache *ollama.ContextCache, systemContext, question string, requestNum int) {
	fmt.Printf("\n--- Request %d: %s ---\n", requestNum, question)

	messages := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeSystem,
			Parts: []llms.ContentPart{
				llms.TextPart(systemContext),
			},
		},
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextPart(question),
			},
		},
	}

	startTime := time.Now()
	resp, err := llm.GenerateContent(ctx, messages,
		llms.WithMaxTokens(150),
		ollama.WithContextCache(cache), // Use context cache
	)
	elapsed := time.Since(startTime)

	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	if len(resp.Choices) > 0 {
		// Show truncated response
		content := resp.Choices[0].Content
		if len(content) > 200 {
			content = content[:200] + "..."
		}
		fmt.Printf("Response: %s\n", content)

		// Check cache status
		if genInfo := resp.Choices[0].GenerationInfo; genInfo != nil {
			if hit, ok := genInfo["CacheHit"].(bool); ok {
				if hit {
					fmt.Printf("✓ Cache HIT - ")
					if cached, ok := genInfo["CachedTokens"].(int); ok {
						fmt.Printf("reused %d tokens", cached)
					}
				} else {
					fmt.Printf("✗ Cache MISS - context stored for reuse")
				}
				fmt.Println()
			}

			if tokens, ok := genInfo["PromptTokens"].(int); ok {
				fmt.Printf("Prompt tokens: %d, ", tokens)
			}
			fmt.Printf("Response time: %v\n", elapsed)
		}
	}
}
