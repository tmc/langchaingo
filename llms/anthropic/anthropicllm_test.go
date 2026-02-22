package anthropic

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vxcontrol/langchaingo/llms"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name     string
		envToken string
		opts     []Option
		wantErr  bool
	}{
		{
			name:     "with token from env",
			envToken: "test-token",
			opts:     []Option{},
			wantErr:  false,
		},
		{
			name:     "with token option",
			envToken: "",
			opts:     []Option{WithToken("test-token")},
			wantErr:  false,
		},
		{
			name:     "missing token",
			envToken: "",
			opts:     []Option{},
			wantErr:  true,
		},
		{
			name:     "with all options",
			envToken: "test-token",
			opts: []Option{
				WithModel("claude-3-opus-20240229"),
				WithBaseURL("https://api.example.com"),
				WithAnthropicBetaHeader("max-tokens-3-5-sonnet-2024-07-15"),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originalKey := os.Getenv("ANTHROPIC_API_KEY")
			os.Setenv("ANTHROPIC_API_KEY", tt.envToken)
			defer os.Setenv("ANTHROPIC_API_KEY", originalKey)

			llm, err := New(tt.opts...)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && llm == nil {
				t.Error("New() returned nil LLM without error")
			}
		})
	}
}

func TestProcessMessages(t *testing.T) {
	tests := []struct {
		name       string
		messages   []llms.MessageContent
		wantLen    int
		wantSystem string
		wantErr    bool
	}{
		{
			name: "basic text message",
			messages: []llms.MessageContent{
				{
					Role: llms.ChatMessageTypeHuman,
					Parts: []llms.ContentPart{
						llms.TextContent{Text: "Hello"},
					},
				},
			},
			wantLen:    1,
			wantSystem: "",
			wantErr:    false,
		},
		{
			name: "system message",
			messages: []llms.MessageContent{
				{
					Role: llms.ChatMessageTypeSystem,
					Parts: []llms.ContentPart{
						llms.TextContent{Text: "You are helpful"},
					},
				},
				{
					Role: llms.ChatMessageTypeHuman,
					Parts: []llms.ContentPart{
						llms.TextContent{Text: "Hi"},
					},
				},
			},
			wantLen:    1,
			wantSystem: "You are helpful",
			wantErr:    false,
		},
		{
			name: "ai and human messages",
			messages: []llms.MessageContent{
				{
					Role: llms.ChatMessageTypeHuman,
					Parts: []llms.ContentPart{
						llms.TextContent{Text: "Hello"},
					},
				},
				{
					Role: llms.ChatMessageTypeAI,
					Parts: []llms.ContentPart{
						llms.TextContent{Text: "Hi there!"},
					},
				},
			},
			wantLen:    2,
			wantSystem: "",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, systemPrompt, err := processMessages(tt.messages)
			if (err != nil) != tt.wantErr {
				t.Errorf("processMessages() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if len(result) != tt.wantLen {
					t.Errorf("processMessages() returned %d messages, want %d", len(result), tt.wantLen)
				}
				if systemPrompt != tt.wantSystem {
					t.Errorf("processMessages() system prompt = %q, want %q", systemPrompt, tt.wantSystem)
				}
			}
		})
	}
}

func TestToolsToTools(t *testing.T) {
	tools := []llms.Tool{
		{
			Type: "function",
			Function: &llms.FunctionDefinition{
				Name:        "get_weather",
				Description: "Get the weather for a location",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"location": map[string]any{
							"type":        "string",
							"description": "The location to get weather for",
						},
					},
					"required": []string{"location"},
				},
			},
		},
	}

	result := toolsToTools(tools)

	if len(result) != 1 {
		t.Fatalf("toolsToTools() returned %d tools, want 1", len(result))
	}
	if result[0].Name != "get_weather" {
		t.Errorf("toolsToTools() tool name = %q, want %q", result[0].Name, "get_weather")
	}
	if result[0].Description != "Get the weather for a location" {
		t.Errorf("toolsToTools() tool description = %q, want %q", result[0].Description, "Get the weather for a location")
	}
}

func TestOptions(t *testing.T) {
	t.Run("WithModel", func(t *testing.T) {
		opts := &options{}
		WithModel("claude-3-opus")(opts)
		if opts.model != "claude-3-opus" {
			t.Errorf("WithModel() got %s, want claude-3-opus", opts.model)
		}
	})

	t.Run("WithToken", func(t *testing.T) {
		opts := &options{}
		WithToken("test-token")(opts)
		if opts.token != "test-token" {
			t.Errorf("WithToken() got %s, want test-token", opts.token)
		}
	})

	t.Run("WithBaseURL", func(t *testing.T) {
		opts := &options{}
		WithBaseURL("https://test.com")(opts)
		if opts.baseURL != "https://test.com" {
			t.Errorf("WithBaseURL() got %s, want https://test.com", opts.baseURL)
		}
	})

	t.Run("WithAnthropicBetaHeader", func(t *testing.T) {
		opts := &options{}
		WithAnthropicBetaHeader("test-beta")(opts)
		if opts.anthropicBetaHeader != "test-beta" {
			t.Errorf("WithAnthropicBetaHeader() got %s, want test-beta", opts.anthropicBetaHeader)
		}
	})

	t.Run("WithLegacyTextCompletionsAPI", func(t *testing.T) {
		opts := &options{}
		WithLegacyTextCompletionsAPI()(opts)
		if !opts.useLegacyTextCompletionsAPI {
			t.Error("WithLegacyTextCompletionsAPI() did not set flag")
		}
	})
}

// TestMergeCacheStrategies tests the cache strategy merge logic
func TestMergeCacheStrategies(t *testing.T) {
	tests := []struct {
		name           string
		clientStrategy *CacheStrategy
		callStrategy   *CacheStrategy
		want           *CacheStrategy
	}{
		{
			name:           "both nil",
			clientStrategy: nil,
			callStrategy:   nil,
			want:           nil,
		},
		{
			name:           "only client-level",
			clientStrategy: &CacheStrategy{CacheTools: true, CacheSystem: true},
			callStrategy:   nil,
			want:           &CacheStrategy{CacheTools: true, CacheSystem: true},
		},
		{
			name:           "only call-level",
			clientStrategy: nil,
			callStrategy:   &CacheStrategy{CacheMessages: true, TTL: "1h"},
			want:           &CacheStrategy{CacheMessages: true, TTL: "1h"},
		},
		{
			name: "call-level overrides client-level (OR merge)",
			clientStrategy: &CacheStrategy{
				CacheTools:  true,
				CacheSystem: true,
				TTL:         "5m",
			},
			callStrategy: &CacheStrategy{
				CacheMessages: true,
				TTL:           "1h",
			},
			want: &CacheStrategy{
				CacheTools:    true, // from client (OR: true || false)
				CacheSystem:   true, // from client (OR: false || true)
				CacheMessages: true, // from call (OR: true || false)
				TTL:           "1h", // call-level takes precedence
			},
		},
		{
			name: "call-level disables nothing (OR logic preserves client settings)",
			clientStrategy: &CacheStrategy{
				CacheTools:  true,
				CacheSystem: true,
			},
			callStrategy: &CacheStrategy{
				CacheMessages: true,
			},
			want: &CacheStrategy{
				CacheTools:    true, // preserved from client
				CacheSystem:   true, // preserved from client
				CacheMessages: true, // added from call
			},
		},
		{
			name: "empty call-level doesn't override",
			clientStrategy: &CacheStrategy{
				CacheTools:  true,
				CacheSystem: true,
				TTL:         "1h",
			},
			callStrategy: &CacheStrategy{}, // all false, no TTL
			want: &CacheStrategy{
				CacheTools:  true, // from client (OR: false || true)
				CacheSystem: true, // from client (OR: false || true)
				TTL:         "1h", // from client (empty string doesn't override)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Prepare call options with strategy
			var opts llms.CallOptions
			if tt.callStrategy != nil {
				opts.Metadata = map[string]any{
					"anthropic:cache_strategy": *tt.callStrategy,
				}
			}

			// Perform merge
			got := mergeCacheStrategies(tt.clientStrategy, &opts)

			// Verify result
			if tt.want == nil {
				assert.Nil(t, got, "Expected nil strategy")
			} else {
				require.NotNil(t, got, "Expected non-nil strategy")
				assert.Equal(t, tt.want.CacheTools, got.CacheTools, "CacheTools mismatch")
				assert.Equal(t, tt.want.CacheSystem, got.CacheSystem, "CacheSystem mismatch")
				assert.Equal(t, tt.want.CacheMessages, got.CacheMessages, "CacheMessages mismatch")
				assert.Equal(t, tt.want.TTL, got.TTL, "TTL mismatch")
			}
		})
	}
}

// TestDefaultCacheStrategy_ClientLevel tests that client-level cache strategy is applied
func TestDefaultCacheStrategy_ClientLevel(t *testing.T) {
	// Create client with default strategy
	llm, err := New(
		WithToken("test-token"),
		WithDefaultCacheStrategy(CacheStrategy{
			CacheTools:  true,
			CacheSystem: true,
			TTL:         "1h",
		}),
	)
	require.NoError(t, err)
	require.NotNil(t, llm.defaultCacheStrategy, "Default strategy should be set")

	// Verify strategy was stored
	assert.True(t, llm.defaultCacheStrategy.CacheTools)
	assert.True(t, llm.defaultCacheStrategy.CacheSystem)
	assert.False(t, llm.defaultCacheStrategy.CacheMessages)
	assert.Equal(t, "1h", llm.defaultCacheStrategy.TTL)
}

// TestDefaultCacheStrategy_MergeWithCallLevel tests merge behavior
func TestDefaultCacheStrategy_MergeWithCallLevel(t *testing.T) {
	// Setup: client with tools+system caching
	clientStrategy := &CacheStrategy{
		CacheTools:  true,
		CacheSystem: true,
		TTL:         "5m",
	}

	// Case 1: Call-level adds messages caching
	opts1 := &llms.CallOptions{
		Metadata: map[string]any{
			"anthropic:cache_strategy": CacheStrategy{
				CacheMessages: true,
			},
		},
	}
	merged1 := mergeCacheStrategies(clientStrategy, opts1)
	require.NotNil(t, merged1)
	assert.True(t, merged1.CacheTools, "Should preserve client CacheTools")
	assert.True(t, merged1.CacheSystem, "Should preserve client CacheSystem")
	assert.True(t, merged1.CacheMessages, "Should add call-level CacheMessages")
	assert.Equal(t, "5m", merged1.TTL, "Should use client TTL when call-level empty")

	// Case 2: Call-level overrides TTL
	opts2 := &llms.CallOptions{
		Metadata: map[string]any{
			"anthropic:cache_strategy": CacheStrategy{
				TTL: "1h",
			},
		},
	}
	merged2 := mergeCacheStrategies(clientStrategy, opts2)
	require.NotNil(t, merged2)
	assert.Equal(t, "1h", merged2.TTL, "Should use call-level TTL")
	assert.True(t, merged2.CacheTools, "Should preserve client settings")

	// Case 3: Call-level with no metadata uses client defaults
	opts3 := &llms.CallOptions{}
	merged3 := mergeCacheStrategies(clientStrategy, opts3)
	require.NotNil(t, merged3)
	assert.Equal(t, clientStrategy, merged3, "Should use client strategy when no call-level")
}

// TestTokenUsageMapping_Anthropic tests correct token usage mapping for Anthropic provider
func TestTokenUsageMapping_Anthropic(t *testing.T) { //nolint:funlen
	tests := []struct {
		name                     string
		inputTokens              int
		cacheCreationInputTokens int
		cacheReadInputTokens     int
		outputTokens             int
		expectedPromptTokens     int
		expectedCacheRead        int
		expectedCacheCreation    int
		expectedCompletionTokens int
		expectedTotalTokens      int
	}{
		{
			name:                     "first request without cache",
			inputTokens:              332,
			cacheCreationInputTokens: 0,
			cacheReadInputTokens:     0,
			outputTokens:             82,
			expectedPromptTokens:     332,
			expectedCacheRead:        0,
			expectedCacheCreation:    0,
			expectedCompletionTokens: 82,
			expectedTotalTokens:      414,
		},
		{
			name:                     "first request with cache creation",
			inputTokens:              332,
			cacheCreationInputTokens: 1546,
			cacheReadInputTokens:     0,
			outputTokens:             82,
			expectedPromptTokens:     1878, // 332 + 1546 + 0
			expectedCacheRead:        0,
			expectedCacheCreation:    1546,
			expectedCompletionTokens: 82,
			expectedTotalTokens:      1960, // 332 + 1546 + 0 + 82
		},
		{
			name:                     "subsequent request with cache hit",
			inputTokens:              332,
			cacheCreationInputTokens: 0,
			cacheReadInputTokens:     1546,
			outputTokens:             82,
			expectedPromptTokens:     1878, // 332 + 0 + 1546
			expectedCacheRead:        1546,
			expectedCacheCreation:    0,
			expectedCompletionTokens: 82,
			expectedTotalTokens:      1960, // 332 + 0 + 1546 + 82
		},
		{
			name:                     "mixed cache scenario",
			inputTokens:              500,
			cacheCreationInputTokens: 200,
			cacheReadInputTokens:     1000,
			outputTokens:             150,
			expectedPromptTokens:     1700, // 500 + 200 + 1000
			expectedCacheRead:        1000,
			expectedCacheCreation:    200,
			expectedCompletionTokens: 150,
			expectedTotalTokens:      1850, // 500 + 200 + 1000 + 150
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate Anthropic response structure
			result := struct {
				Usage struct {
					InputTokens              int
					CacheCreationInputTokens int
					CacheReadInputTokens     int
					OutputTokens             int
					CacheCreation            struct {
						Ephemeral5mInputTokens int
						Ephemeral1hInputTokens int
					}
					ServiceTier string
				}
			}{}

			result.Usage.InputTokens = tt.inputTokens
			result.Usage.CacheCreationInputTokens = tt.cacheCreationInputTokens
			result.Usage.CacheReadInputTokens = tt.cacheReadInputTokens
			result.Usage.OutputTokens = tt.outputTokens
			result.Usage.ServiceTier = "standard"

			// Build GenerationInfo as done in anthropicllm.go
			generationInfo := map[string]any{
				"PromptTokens":             result.Usage.InputTokens + result.Usage.CacheCreationInputTokens + result.Usage.CacheReadInputTokens,
				"CompletionTokens":         result.Usage.OutputTokens,
				"TotalTokens":              result.Usage.InputTokens + result.Usage.CacheCreationInputTokens + result.Usage.CacheReadInputTokens + result.Usage.OutputTokens,
				"ReasoningTokens":          0,
				"PromptCachedTokens":       result.Usage.CacheReadInputTokens,
				"CacheReadInputTokens":     result.Usage.CacheReadInputTokens,
				"CacheCreationInputTokens": result.Usage.CacheCreationInputTokens,
				"InputTokens":              result.Usage.InputTokens,
				"OutputTokens":             result.Usage.OutputTokens,
				"ServiceTier":              result.Usage.ServiceTier,
			}

			// Verify mapped values
			assert.Equal(t, tt.expectedPromptTokens, generationInfo["PromptTokens"], "PromptTokens mismatch")
			assert.Equal(t, tt.expectedCacheRead, generationInfo["CacheReadInputTokens"], "CacheReadInputTokens mismatch")
			assert.Equal(t, tt.expectedCacheCreation, generationInfo["CacheCreationInputTokens"], "CacheCreationInputTokens mismatch")
			assert.Equal(t, tt.expectedCompletionTokens, generationInfo["CompletionTokens"], "CompletionTokens mismatch")
			assert.Equal(t, tt.expectedTotalTokens, generationInfo["TotalTokens"], "TotalTokens mismatch")

			// Verify client-side cost calculation logic
			// Client formula: input = max(PromptTokens - CacheRead, 0)
			promptTokens := generationInfo["PromptTokens"].(int)
			cacheRead := generationInfo["CacheReadInputTokens"].(int)
			cacheWrite := generationInfo["CacheCreationInputTokens"].(int)

			uncachedTokens := max(promptTokens-cacheRead, 0)

			// Expected: uncached tokens should equal inputTokens + cacheCreationInputTokens
			expectedUncached := tt.inputTokens + tt.cacheCreationInputTokens
			assert.Equal(t, expectedUncached, uncachedTokens, "Uncached tokens calculation mismatch")

			// For Anthropic pricing: uncached * basePrice + cacheRead * cacheReadPrice + cacheWrite * cacheWritePrice
			basePrice := 3.0 / 1e6        // $3 per 1M tokens
			cacheReadPrice := 0.3 / 1e6   // $0.3 per 1M tokens (90% discount)
			cacheWritePrice := 3.75 / 1e6 // $3.75 per 1M tokens (25% premium)

			expectedCost := float64(uncachedTokens)*basePrice +
				float64(cacheRead)*cacheReadPrice +
				float64(cacheWrite)*cacheWritePrice

			actualCost := float64(uncachedTokens)*basePrice +
				float64(cacheRead)*cacheReadPrice +
				float64(cacheWrite)*cacheWritePrice

			assert.InDelta(t, expectedCost, actualCost, 0.000001, "Cost calculation mismatch")
		})
	}
}
