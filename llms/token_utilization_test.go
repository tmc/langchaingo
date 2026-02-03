package llms_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/vxcontrol/langchaingo/llms"
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

// ClientCallUsage represents client-side token usage tracking
// This mirrors the structure used in client code for cost calculation
type ClientCallUsage struct {
	Input      int64
	Output     int64
	CacheRead  int64
	CacheWrite int64
	CostInput  float64
	CostOutput float64
}

// ClientPriceInfo represents pricing information for a model
type ClientPriceInfo struct {
	Input      float64
	Output     float64
	CacheRead  float64
	CacheWrite float64
}

// UpdateCost calculates the cost based on usage and pricing
// This is the client-side formula that needs to work for all providers
func (c *ClientCallUsage) UpdateCost(price *ClientPriceInfo) {
	if price == nil {
		return
	}

	// If cost already calculated (e.g., by OpenRouter), don't override
	if c.CostInput != 0.0 || c.CostOutput != 0.0 {
		return
	}

	// If no cache pricing, calculate as if cache is not used
	if price.CacheRead == 0.0 && price.CacheWrite == 0.0 {
		c.CostInput = float64(c.Input) * price.Input / 1e6
		c.CostOutput = float64(c.Output) * price.Output / 1e6
		return
	}

	// Calculate with cache pricing
	input := max(float64(c.Input-c.CacheRead), 0.0)
	output := float64(c.Output)
	cacheReadCost := float64(c.CacheRead) * price.CacheRead / 1e6
	cacheWriteCost := float64(c.CacheWrite) * price.CacheWrite / 1e6
	c.CostInput = input*price.Input/1e6 + cacheReadCost + cacheWriteCost
	c.CostOutput = output * price.Output / 1e6
}

// TestClientCostCalculation_Anthropic tests client-side cost calculation for Anthropic
func TestClientCostCalculation_Anthropic(t *testing.T) {
	tests := []struct {
		name              string
		promptTokens      int
		cacheRead         int
		cacheWrite        int
		completionTokens  int
		withCachePrice    bool
		expectedInputCost float64
	}{
		{
			name:              "first request with cache creation",
			promptTokens:      1878, // 332 + 1546 + 0
			cacheRead:         0,
			cacheWrite:        1546,
			completionTokens:  82,
			withCachePrice:    true,
			expectedInputCost: (332+1546)*3.0/1e6 + 0*0.3/1e6 + 1546*3.75/1e6,
		},
		{
			name:              "subsequent request with cache hit",
			promptTokens:      1878, // 332 + 0 + 1546
			cacheRead:         1546,
			cacheWrite:        0,
			completionTokens:  82,
			withCachePrice:    true,
			expectedInputCost: 332*3.0/1e6 + 1546*0.3/1e6 + 0*3.75/1e6,
		},
		{
			name:              "without cache pricing (fallback)",
			promptTokens:      1878,
			cacheRead:         1546,
			cacheWrite:        0,
			completionTokens:  82,
			withCachePrice:    false,
			expectedInputCost: 1878 * 3.0 / 1e6, // All at base price
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			usage := ClientCallUsage{
				Input:      int64(tt.promptTokens),
				Output:     int64(tt.completionTokens),
				CacheRead:  int64(tt.cacheRead),
				CacheWrite: int64(tt.cacheWrite),
			}

			var price *ClientPriceInfo
			if tt.withCachePrice {
				price = &ClientPriceInfo{
					Input:      3.0,  // $3 per 1M tokens
					Output:     15.0, // $15 per 1M tokens
					CacheRead:  0.3,  // $0.3 per 1M tokens
					CacheWrite: 3.75, // $3.75 per 1M tokens
				}
			} else {
				price = &ClientPriceInfo{
					Input:  3.0,
					Output: 15.0,
					// CacheRead and CacheWrite are 0
				}
			}

			usage.UpdateCost(price)

			if fmt.Sprintf("%.9f", usage.CostInput) != fmt.Sprintf("%.9f", tt.expectedInputCost) {
				t.Errorf("Input cost mismatch: expected %.9f, got %.9f", tt.expectedInputCost, usage.CostInput)
			}
		})
	}
}

// TestClientCostCalculation_OpenAI tests client-side cost calculation for OpenAI
func TestClientCostCalculation_OpenAI(t *testing.T) {
	tests := []struct {
		name              string
		promptTokens      int
		cacheRead         int
		cacheWrite        int
		completionTokens  int
		withCachePrice    bool
		expectedInputCost float64
	}{
		{
			name:              "first request without cache",
			promptTokens:      2619,
			cacheRead:         0,
			cacheWrite:        0,
			completionTokens:  149,
			withCachePrice:    true,
			expectedInputCost: 2619 * 2.5 / 1e6,
		},
		{
			name:              "subsequent request with cache hit",
			promptTokens:      2619,
			cacheRead:         2048,
			cacheWrite:        0,
			completionTokens:  85,
			withCachePrice:    true,
			expectedInputCost: (2619-2048)*2.5/1e6 + 2048*1.25/1e6,
		},
		{
			name:              "without cache pricing (fallback)",
			promptTokens:      2619,
			cacheRead:         2048,
			cacheWrite:        0,
			completionTokens:  85,
			withCachePrice:    false,
			expectedInputCost: 2619 * 2.5 / 1e6, // All at base price
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			usage := ClientCallUsage{
				Input:      int64(tt.promptTokens),
				Output:     int64(tt.completionTokens),
				CacheRead:  int64(tt.cacheRead),
				CacheWrite: int64(tt.cacheWrite),
			}

			var price *ClientPriceInfo
			if tt.withCachePrice {
				price = &ClientPriceInfo{
					Input:     2.5,  // $2.5 per 1M tokens
					Output:    10.0, // $10 per 1M tokens
					CacheRead: 1.25, // $1.25 per 1M tokens (50% discount)
					// CacheWrite is 0 for OpenAI (no extra charge)
				}
			} else {
				price = &ClientPriceInfo{
					Input:  2.5,
					Output: 10.0,
				}
			}

			usage.UpdateCost(price)

			if fmt.Sprintf("%.9f", usage.CostInput) != fmt.Sprintf("%.9f", tt.expectedInputCost) {
				t.Errorf("Input cost mismatch: expected %.9f, got %.9f", tt.expectedInputCost, usage.CostInput)
			}
		})
	}
}

// TestClientCostCalculation_GoogleAI tests client-side cost calculation for Google AI
func TestClientCostCalculation_GoogleAI(t *testing.T) {
	tests := []struct {
		name              string
		promptTokens      int
		cacheRead         int
		cacheWrite        int
		completionTokens  int
		withCachePrice    bool
		expectedInputCost float64
	}{
		{
			name:              "first request without cache",
			promptTokens:      4517,
			cacheRead:         0,
			cacheWrite:        0,
			completionTokens:  9,
			withCachePrice:    true,
			expectedInputCost: 4517 * 0.075 / 1e6,
		},
		{
			name:              "subsequent request with cache hit",
			promptTokens:      4534,
			cacheRead:         4058,
			cacheWrite:        0,
			completionTokens:  11,
			withCachePrice:    true,
			expectedInputCost: (4534-4058)*0.075/1e6 + 4058*0.01875/1e6,
		},
		{
			name:              "without cache pricing (fallback)",
			promptTokens:      4534,
			cacheRead:         4058,
			cacheWrite:        0,
			completionTokens:  11,
			withCachePrice:    false,
			expectedInputCost: 4534 * 0.075 / 1e6, // All at base price
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			usage := ClientCallUsage{
				Input:      int64(tt.promptTokens),
				Output:     int64(tt.completionTokens),
				CacheRead:  int64(tt.cacheRead),
				CacheWrite: int64(tt.cacheWrite),
			}

			var price *ClientPriceInfo
			if tt.withCachePrice {
				price = &ClientPriceInfo{
					Input:     0.075,   // $0.075 per 1M tokens
					Output:    0.30,    // $0.30 per 1M tokens
					CacheRead: 0.01875, // $0.01875 per 1M tokens (25% of base)
					// CacheWrite is 0 for Google AI
				}
			} else {
				price = &ClientPriceInfo{
					Input:  0.075,
					Output: 0.30,
				}
			}

			usage.UpdateCost(price)

			if fmt.Sprintf("%.9f", usage.CostInput) != fmt.Sprintf("%.9f", tt.expectedInputCost) {
				t.Errorf("Input cost mismatch: expected %.9f, got %.9f", tt.expectedInputCost, usage.CostInput)
			}
		})
	}
}

// TestClientCostCalculation_CrossProvider tests that the same client logic works for all providers
func TestClientCostCalculation_CrossProvider(t *testing.T) {
	// This test verifies that the unified client formula works correctly
	// for all three providers with their different token reporting styles

	scenarios := []struct {
		provider         string
		promptTokens     int
		cacheRead        int
		cacheWrite       int
		completionTokens int
		basePrice        float64
		cacheReadPrice   float64
		cacheWritePrice  float64
		expectedUncached int
	}{
		{
			provider:         "Anthropic with cache hit",
			promptTokens:     1878, // Includes all: 332 (uncached) + 0 (write) + 1546 (read)
			cacheRead:        1546,
			cacheWrite:       0,
			completionTokens: 82,
			basePrice:        3.0,
			cacheReadPrice:   0.3,
			cacheWritePrice:  3.75,
			expectedUncached: 332, // 1878 - 1546
		},
		{
			provider:         "OpenAI with cache hit",
			promptTokens:     2619, // Already includes all tokens
			cacheRead:        2048,
			cacheWrite:       0,
			completionTokens: 85,
			basePrice:        2.5,
			cacheReadPrice:   1.25,
			cacheWritePrice:  0,
			expectedUncached: 571, // 2619 - 2048
		},
		{
			provider:         "Google AI with cache hit",
			promptTokens:     4534, // Already includes all tokens
			cacheRead:        4058,
			cacheWrite:       0,
			completionTokens: 11,
			basePrice:        0.075,
			cacheReadPrice:   0.01875,
			cacheWritePrice:  0,
			expectedUncached: 476, // 4534 - 4058
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.provider, func(t *testing.T) {
			usage := ClientCallUsage{
				Input:      int64(scenario.promptTokens),
				Output:     int64(scenario.completionTokens),
				CacheRead:  int64(scenario.cacheRead),
				CacheWrite: int64(scenario.cacheWrite),
			}

			price := &ClientPriceInfo{
				Input:      scenario.basePrice,
				Output:     10.0, // Not important for this test
				CacheRead:  scenario.cacheReadPrice,
				CacheWrite: scenario.cacheWritePrice,
			}

			usage.UpdateCost(price)

			// Verify uncached tokens calculation
			uncached := max(int(usage.Input-usage.CacheRead), 0)
			if uncached != scenario.expectedUncached {
				t.Errorf("Uncached tokens mismatch: expected %d, got %d", scenario.expectedUncached, uncached)
			}

			// Verify cost calculation
			expectedCost := float64(scenario.expectedUncached)*scenario.basePrice/1e6 +
				float64(scenario.cacheRead)*scenario.cacheReadPrice/1e6 +
				float64(scenario.cacheWrite)*scenario.cacheWritePrice/1e6

			if fmt.Sprintf("%.9f", usage.CostInput) != fmt.Sprintf("%.9f", expectedCost) {
				t.Errorf("Cost mismatch: expected %.9f, got %.9f", expectedCost, usage.CostInput)
			}
		})
	}
}
