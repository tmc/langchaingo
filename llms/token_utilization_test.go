package llms_test

import (
	"context"
	"testing"

	"github.com/vendasta/langchaingo/llms"
)

// MockLLMWithTokenUsage is a mock LLM that returns token usage information
type MockLLMWithTokenUsage struct {
	includeCache bool
}

func (m *MockLLMWithTokenUsage) Call(ctx context.Context, prompt string, options ...llms.CallOption) (string, error) {
	return "test response", nil
}

func (m *MockLLMWithTokenUsage) GenerateContent(ctx context.Context, messages []llms.MessageContent, options ...llms.CallOption) (*llms.ContentResponse, error) {
	generationInfo := map[string]any{
		"CompletionTokens": 50,
		"PromptTokens":     100,
		"TotalTokens":      150,
	}

	if m.includeCache {
		// OpenAI-style cache tokens
		generationInfo["PromptCachedTokens"] = 80

		// Anthropic-style cache tokens
		generationInfo["CacheCreationInputTokens"] = 20
		generationInfo["CacheReadInputTokens"] = 80
	}

	return &llms.ContentResponse{
		Choices: []*llms.ContentChoice{
			{
				Content:        "test response",
				GenerationInfo: generationInfo,
			},
		},
	}, nil
}

func TestTokenUtilizationWithoutCache(t *testing.T) {
	llm := &MockLLMWithTokenUsage{includeCache: false}
	ctx := context.Background()

	messages := []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextPart("test")},
		},
	}

	resp, err := llm.GenerateContent(ctx, messages)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(resp.Choices) == 0 {
		t.Fatal("expected at least one choice")
	}

	info := resp.Choices[0].GenerationInfo

	// Check basic token counts
	if ct, ok := info["CompletionTokens"].(int); !ok || ct != 50 {
		t.Errorf("expected CompletionTokens=50, got %v", info["CompletionTokens"])
	}

	if pt, ok := info["PromptTokens"].(int); !ok || pt != 100 {
		t.Errorf("expected PromptTokens=100, got %v", info["PromptTokens"])
	}

	if tt, ok := info["TotalTokens"].(int); !ok || tt != 150 {
		t.Errorf("expected TotalTokens=150, got %v", info["TotalTokens"])
	}

	// Cache tokens should not be present
	if _, ok := info["PromptCachedTokens"]; ok {
		t.Error("PromptCachedTokens should not be present")
	}

	if _, ok := info["CacheCreationInputTokens"]; ok {
		t.Error("CacheCreationInputTokens should not be present")
	}

	if _, ok := info["CacheReadInputTokens"]; ok {
		t.Error("CacheReadInputTokens should not be present")
	}
}

func TestTokenUtilizationWithCache(t *testing.T) {
	llm := &MockLLMWithTokenUsage{includeCache: true}
	ctx := context.Background()

	messages := []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextPart("test")},
		},
	}

	resp, err := llm.GenerateContent(ctx, messages)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(resp.Choices) == 0 {
		t.Fatal("expected at least one choice")
	}

	info := resp.Choices[0].GenerationInfo

	// Check basic token counts
	if ct, ok := info["CompletionTokens"].(int); !ok || ct != 50 {
		t.Errorf("expected CompletionTokens=50, got %v", info["CompletionTokens"])
	}

	if pt, ok := info["PromptTokens"].(int); !ok || pt != 100 {
		t.Errorf("expected PromptTokens=100, got %v", info["PromptTokens"])
	}

	if tt, ok := info["TotalTokens"].(int); !ok || tt != 150 {
		t.Errorf("expected TotalTokens=150, got %v", info["TotalTokens"])
	}

	// OpenAI-style cache tokens
	if pct, ok := info["PromptCachedTokens"].(int); !ok || pct != 80 {
		t.Errorf("expected PromptCachedTokens=80, got %v", info["PromptCachedTokens"])
	}

	// Anthropic-style cache tokens
	if ccit, ok := info["CacheCreationInputTokens"].(int); !ok || ccit != 20 {
		t.Errorf("expected CacheCreationInputTokens=20, got %v", info["CacheCreationInputTokens"])
	}

	if crit, ok := info["CacheReadInputTokens"].(int); !ok || crit != 80 {
		t.Errorf("expected CacheReadInputTokens=80, got %v", info["CacheReadInputTokens"])
	}
}

func TestCalculateCostSavings(t *testing.T) {
	// Test function to calculate cost savings from cached tokens
	tests := []struct {
		name            string
		promptTokens    int
		cachedTokens    int
		pricePerMToken  float64
		expectedSavings float64
	}{
		{
			name:            "OpenAI 50% discount",
			promptTokens:    1000,
			cachedTokens:    800,
			pricePerMToken:  5.0,   // $5 per 1M tokens
			expectedSavings: 0.002, // 800 tokens * 50% discount * $5/1M
		},
		{
			name:            "Anthropic 90% discount",
			promptTokens:    2000,
			cachedTokens:    1500,
			pricePerMToken:  15.0,    // $15 per 1M tokens
			expectedSavings: 0.02025, // 1500 tokens * 90% discount * $15/1M
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Calculate savings
			var discountRate float64
			if tt.name == "OpenAI 50% discount" {
				discountRate = 0.5
			} else {
				discountRate = 0.9
			}

			savings := float64(tt.cachedTokens) * discountRate * tt.pricePerMToken / 1_000_000

			if savings != tt.expectedSavings {
				t.Errorf("expected savings=%f, got %f", tt.expectedSavings, savings)
			}
		})
	}
}
