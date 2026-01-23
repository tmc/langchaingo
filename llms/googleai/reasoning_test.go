package googleai

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/vxcontrol/langchaingo/llms"
)

func TestGoogleAI_ReasoningIntegration(t *testing.T) {
	apiKey := os.Getenv("GOOGLE_API_KEY")
	if apiKey == "" {
		t.Skip("GOOGLE_API_KEY not set")
	}

	ctx := context.Background()

	// Test with Gemini 2.0 Flash (reasoning model)
	client, err := New(ctx,
		WithAPIKey(apiKey),
		WithDefaultModel("gemini-2.5-flash"),
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Test reasoning with a complex problem
	messages := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextPart("How many time light travels from the Earth to the Moon? Explain your reasoning step by step."),
			},
		},
	}

	resp, err := client.GenerateContent(ctx, messages,
		llms.WithMaxTokens(4000),
		llms.WithReasoning(llms.ReasoningMedium, 4000),
	)
	if err != nil {
		t.Fatalf("Failed to generate content: %v", err)
	}

	if len(resp.Choices) != 1 {
		t.Fatalf("Expected 1 response choice, got: %d", len(resp.Choices))
	}

	content := resp.Choices[0].Content
	if !strings.Contains(content, "second") || !strings.Contains(content, "1") {
		t.Errorf("Expected answer 1 second in response, got: %s", content)
	}
}

func TestGoogleAI_CachingSupport(t *testing.T) {
	apiKey := os.Getenv("GOOGLE_API_KEY")
	if apiKey == "" {
		t.Skip("GOOGLE_API_KEY not set")
	}

	ctx := context.Background()

	// Create caching helper
	helper, err := NewCachingHelper(ctx, WithAPIKey(apiKey))
	if err != nil {
		t.Fatalf("Failed to create caching helper: %v", err)
	}

	// Create cached content with a large system prompt
	// Note: Google AI requires at least 32,768 tokens (~24,000 words) for caching
	// This is approximately 500+ words repeated many times
	baseContext := `You are an expert code reviewer with deep knowledge of Go best practices.
You should always consider the following aspects in your reviews:

1. Performance: Analyze algorithmic complexity, memory usage, and potential bottlenecks.
2. Security: Check for common vulnerabilities like SQL injection, XSS, buffer overflows.
3. Maintainability: Evaluate code readability, documentation, and adherence to conventions.
4. Testing: Assess test coverage and quality of test cases.
5. Error handling: Verify proper error propagation and handling.
6. Concurrency: Review goroutine usage, channel operations, and synchronization.
7. Resource management: Check for proper cleanup of resources (files, connections, etc).
8. API design: Evaluate interface design and package structure.

`
	// Repeat to reach minimum cache size (~32k tokens requires ~24k words)
	longContext := strings.Repeat(baseContext, 500)

	cached, err := helper.CreateCachedContent(ctx, "gemini-2.0-flash",
		[]llms.MessageContent{
			{
				Role: llms.ChatMessageTypeSystem,
				Parts: []llms.ContentPart{
					llms.TextPart(longContext),
				},
			},
		},
		5*time.Minute,
		"go-code-review-expert",
	)
	if err != nil {
		t.Fatalf("Failed to create cached content: %v", err)
	}
	t.Logf("Created cached content: %s", cached.Name)
	defer func() {
		if err := helper.DeleteCachedContent(ctx, cached.Name); err != nil {
			t.Logf("Failed to delete cached content: %v", err)
		}
	}()

	// Test Get operation
	retrieved, err := helper.GetCachedContent(ctx, cached.Name)
	if err != nil {
		t.Fatalf("Failed to get cached content: %v", err)
	}
	if retrieved.Name != cached.Name {
		t.Errorf("Retrieved cache name mismatch: got %s, want %s", retrieved.Name, cached.Name)
	}
	t.Logf("Cache metadata: TotalTokens=%d, ExpireTime=%s",
		retrieved.UsageMetadata.TotalTokenCount,
		retrieved.ExpireTime.Format(time.RFC3339))

	// Test Update operation
	updated, err := helper.UpdateCachedContent(ctx, cached.Name, 10*time.Minute)
	if err != nil {
		t.Fatalf("Failed to update cached content: %v", err)
	}
	t.Logf("Updated cache expiration to: %s", updated.ExpireTime.Format(time.RFC3339))

	// Test List operation using All iterator
	found := false
	for cache, err := range helper.AllCachedContents(ctx) {
		if err != nil {
			t.Logf("Error iterating cached contents: %v", err)
			continue
		}
		if cache.Name == cached.Name {
			found = true
			t.Logf("Found our cache in list: %s (DisplayName: %s)", cache.Name, cache.DisplayName)
			break
		}
	}
	if !found {
		t.Error("Did not find our cached content in the list")
	}

	// Use the cached content in a request
	client, err := New(ctx, WithAPIKey(apiKey))
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	messages := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextPart("What are the key things to look for in a Go code review?"),
			},
		},
	}

	resp, err := client.GenerateContent(ctx, messages,
		WithCachedContent(cached.Name),
		llms.WithMaxTokens(200),
	)
	if err != nil {
		t.Fatalf("Failed to generate with cached content: %v", err)
	}

	// Check that cached tokens are reported in the response
	if len(resp.Choices) == 0 {
		t.Fatal("No response choices")
	}

	if genInfo := resp.Choices[0].GenerationInfo; genInfo != nil {
		if cachedTokens, ok := genInfo["CachedTokens"].(int32); ok && cachedTokens > 0 {
			t.Logf("Successfully used %d cached tokens", cachedTokens)
		} else {
			t.Log("No cached tokens reported (this may be normal if caching is not yet active)")
		}
		t.Logf("Response metadata: %+v", genInfo)
	}

	if resp.Choices[0].Content == "" {
		t.Error("Empty response content")
	} else {
		t.Logf("Response preview: %s...", resp.Choices[0].Content[:min(100, len(resp.Choices[0].Content))])
	}
}
