package googleaiv2

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/vendasta/langchaingo/llms"
)

func TestGoogleAI_SupportsReasoning(t *testing.T) {
	// Note: SupportsReasoning returns true only for models that support ThinkingConfig API
	// This includes models with "thinking" in the name and gemini-3+ models
	// Regular gemini-2.0-flash does NOT support ThinkingConfig (only gemini-2.0-flash-thinking-exp does)
	tests := []struct {
		name     string
		model    string
		expected bool
	}{
		{
			name:     "Gemini 2.0 Flash does NOT support ThinkingConfig",
			model:    "gemini-2.0-flash",
			expected: false,
		},
		{
			name:     "Gemini 2.0 Pro does NOT support ThinkingConfig",
			model:    "gemini-2.0-pro",
			expected: false,
		},
		{
			name:     "Gemini 2.0 Flash Thinking Exp supports ThinkingConfig",
			model:    "gemini-2.0-flash-thinking-exp",
			expected: true,
		},
		{
			name:     "Gemini 3.0 (future) supports reasoning",
			model:    "gemini-3.0-ultra",
			expected: true,
		},
		{
			name:     "Gemini 3.0 Pro Preview supports reasoning",
			model:    "gemini-3-pro-preview",
			expected: true,
		},
		{
			name:     "Gemini experimental with thinking supports reasoning",
			model:    "gemini-exp-thinking-1206",
			expected: true,
		},
		{
			name:     "Gemini 1.5 Flash does not support reasoning",
			model:    "gemini-1.5-flash",
			expected: false,
		},
		{
			name:     "Gemini 1.0 Pro does not support reasoning",
			model:    "gemini-1.0-pro",
			expected: false,
		},
		{
			name:     "Gemini experimental without thinking does not support reasoning",
			model:    "gemini-exp-1206",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create client with test model
			client := &GoogleAI{
				model: tt.model,
				opts:  DefaultOptions(),
			}

			// Test SupportsReasoning
			got := client.SupportsReasoning()
			if got != tt.expected {
				t.Errorf("SupportsReasoning() for model %s = %v, want %v", tt.model, got, tt.expected)
			}

			// Also test with model set via options
			client.model = ""
			client.opts.DefaultModel = tt.model
			got = client.SupportsReasoning()
			if got != tt.expected {
				t.Errorf("SupportsReasoning() with default model %s = %v, want %v", tt.model, got, tt.expected)
			}
		})
	}
}

func TestGoogleAI_ReasoningIntegration(t *testing.T) {
	apiKey := os.Getenv("GOOGLE_API_KEY")
	if apiKey == "" {
		t.Skip("GOOGLE_API_KEY not set")
	}

	ctx := context.Background()

	// Test with Gemini 2.0 Flash - this model can do reasoning prompts but
	// does NOT support ThinkingConfig API (only thinking-exp models do)
	client, err := New(ctx,
		WithAPIKey(apiKey),
		WithDefaultModel("gemini-2.0-flash"),
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	// Verify it implements ReasoningModel interface
	if _, ok := interface{}(client).(llms.ReasoningModel); !ok {
		t.Error("GoogleAI should implement ReasoningModel interface")
	}

	// Note: gemini-2.0-flash does NOT support ThinkingConfig API
	// Only gemini-2.0-flash-thinking-exp and gemini-3+ models support it
	if client.SupportsReasoning() {
		t.Error("Gemini 2.0 Flash should NOT report SupportsReasoning (use thinking-exp models)")
	}

	// Test that the model can still handle reasoning prompts (just without ThinkingConfig)
	messages := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextPart("What is 25 * 17 + 13? Think step by step."),
			},
		},
	}

	// Test without ThinkingMode since gemini-2.0-flash doesn't support it
	resp, err := client.GenerateContent(ctx, messages,
		llms.WithMaxTokens(200),
	)
	if err != nil {
		t.Fatalf("Failed to generate content: %v", err)
	}

	if len(resp.Choices) == 0 {
		t.Fatal("No response choices")
	}

	content := resp.Choices[0].Content
	if !strings.Contains(content, "438") {
		t.Errorf("Expected answer 438 in response, got: %s", content)
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
	// Google AI requires at least 4096 tokens for cached content
	// Each repetition is approximately 10 tokens, so we need ~400+ repetitions
	longContext := `You are an expert code reviewer with deep knowledge of Go best practices.
	Always consider performance, security, and maintainability in your reviews.
	` + strings.Repeat("This is padding text to ensure we have enough tokens for caching. ", 500)

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
	)
	if err != nil {
		t.Fatalf("Failed to create cached content: %v", err)
	}
	defer func() {
		if err := helper.DeleteCachedContent(ctx, cached.Name); err != nil {
			t.Logf("Failed to delete cached content: %v", err)
		}
	}()

	// Use the cached content in a request
	client, err := New(ctx, WithAPIKey(apiKey), WithRest())
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

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
	if genInfo := resp.Choices[0].GenerationInfo; genInfo != nil {
		if cachedTokens, ok := genInfo["CachedTokens"].(int32); ok && cachedTokens > 0 {
			t.Logf("Successfully used %d cached tokens", cachedTokens)
		} else {
			t.Log("No cached tokens reported (this may be normal if caching is not yet active)")
		}
	}
}
