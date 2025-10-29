package anthropic

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/anthropic/internal/anthropicclient"
)

func TestExtractThinkingOptions(t *testing.T) {
	t.Parallel()

	llm := &LLM{
		model: "claude-sonnet-4-20250514",
	}

	tests := []struct {
		name                 string
		reasoning            *llms.ReasoningOptions
		wantBudget           int
		wantInterleaved      bool
		wantExtendedThinking bool
	}{
		{
			name:                 "Low thinking mode",
			reasoning:            &llms.ReasoningOptions{Mode: llms.ThinkingModeLow},
			wantBudget:           2000,
			wantInterleaved:      false,
			wantExtendedThinking: true,
		},
		{
			name:                 "Medium thinking mode",
			reasoning:            &llms.ReasoningOptions{Mode: llms.ThinkingModeMedium},
			wantBudget:           8000,
			wantInterleaved:      false,
			wantExtendedThinking: true,
		},
		{
			name:                 "High thinking mode",
			reasoning:            &llms.ReasoningOptions{Mode: llms.ThinkingModeHigh},
			wantBudget:           16000,
			wantInterleaved:      false,
			wantExtendedThinking: true,
		},
		{
			name: "Custom budget",
			reasoning: &llms.ReasoningOptions{
				BudgetTokens: ptrInt(5000),
			},
			wantBudget:           5000,
			wantInterleaved:      false,
			wantExtendedThinking: true,
		},
		{
			name: "Interleaved thinking",
			reasoning: &llms.ReasoningOptions{
				Mode:        llms.ThinkingModeMedium,
				Interleaved: true,
			},
			wantBudget:           8000,
			wantInterleaved:      true,
			wantExtendedThinking: true,
		},
		{
			name: "Budget below minimum",
			reasoning: &llms.ReasoningOptions{
				BudgetTokens: ptrInt(500), // Below 1024 minimum
			},
			wantBudget:           1024, // Should be clamped to minimum
			wantInterleaved:      false,
			wantExtendedThinking: true,
		},
		{
			name: "Budget above maximum",
			reasoning: &llms.ReasoningOptions{
				BudgetTokens: ptrInt(200000), // Above 128K maximum
			},
			wantBudget:           128000, // Should be clamped to maximum
			wantInterleaved:      false,
			wantExtendedThinking: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			opts := &llms.CallOptions{
				Reasoning: tt.reasoning,
			}

			betaHeaders, thinking := extractThinkingOptions(llm, opts)

			// Check thinking config
			if tt.wantBudget > 0 {
				require.NotNil(t, thinking, "thinking config should not be nil")
				assert.Equal(t, "enabled", thinking.Type)
				assert.Equal(t, tt.wantBudget, thinking.BudgetTokens)
			}

			// Check beta headers
			hasExtendedThinking := false
			hasInterleaved := false
			for _, header := range betaHeaders {
				if header == "extended-thinking-2025-01-01" {
					hasExtendedThinking = true
				}
				if header == "interleaved-thinking-2025-05-14" {
					hasInterleaved = true
				}
			}

			assert.Equal(t, tt.wantExtendedThinking, hasExtendedThinking, "extended thinking header")
			assert.Equal(t, tt.wantInterleaved, hasInterleaved, "interleaved thinking header")
		})
	}
}

func TestExtractThinkingOptionsUnsupportedModel(t *testing.T) {
	t.Parallel()

	llm := &LLM{
		model: "claude-3-haiku-20240307", // Old model without thinking support
	}

	opts := &llms.CallOptions{
		Reasoning: &llms.ReasoningOptions{
			Mode: llms.ThinkingModeMedium,
		},
	}

	betaHeaders, thinking := extractThinkingOptions(llm, opts)

	// Should not configure thinking for unsupported models
	assert.Nil(t, thinking, "thinking should be nil for unsupported model")
	assert.Empty(t, betaHeaders, "no beta headers should be added for unsupported model")
}

func TestTemperatureEnforcement(t *testing.T) {
	t.Parallel()

	// Test that temperature is enforced to 1.0 when thinking is enabled
	// This is tested indirectly through the generateMessagesContent function
	// We verify this in integration tests
}

func TestProcessAnthropicResponseWithThinking(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		payload         *anthropicclient.MessageResponsePayload
		thinkingConfig  *anthropicclient.ThinkingConfig
		wantContent     string
		wantThinking    string
		wantTokens      int
		wantBudgetAlloc int
		wantBudgetUsed  int
	}{
		{
			name: "Response with thinking content",
			payload: &anthropicclient.MessageResponsePayload{
				Content: []anthropicclient.Content{
					&anthropicclient.ThinkingContent{
						Type:     "thinking",
						Thinking: "Let me analyze this problem step by step...",
					},
					&anthropicclient.TextContent{
						Type: "text",
						Text: "The answer is 42.",
					},
				},
				StopReason: "end_turn",
			},
			thinkingConfig: &anthropicclient.ThinkingConfig{
				Type:         "enabled",
				BudgetTokens: 8000,
			},
			wantContent:     "The answer is 42.",
			wantThinking:    "Let me analyze this problem step by step...",
			wantTokens:      11, // Approximate: ~44 chars / 4
			wantBudgetAlloc: 8000,
			wantBudgetUsed:  11,
		},
		{
			name: "Response without thinking",
			payload: &anthropicclient.MessageResponsePayload{
				Content: []anthropicclient.Content{
					&anthropicclient.TextContent{
						Type: "text",
						Text: "Hello, world!",
					},
				},
				StopReason: "end_turn",
			},
			thinkingConfig:  nil,
			wantContent:     "Hello, world!",
			wantThinking:    "",
			wantTokens:      0,
			wantBudgetAlloc: 0,
			wantBudgetUsed:  0,
		},
		{
			name: "Multiple thinking blocks",
			payload: &anthropicclient.MessageResponsePayload{
				Content: []anthropicclient.Content{
					&anthropicclient.ThinkingContent{
						Type:     "thinking",
						Thinking: "First thought...",
					},
					&anthropicclient.ThinkingContent{
						Type:     "thinking",
						Thinking: "Second thought...",
					},
					&anthropicclient.TextContent{
						Type: "text",
						Text: "Final answer.",
					},
				},
				StopReason: "end_turn",
			},
			thinkingConfig: &anthropicclient.ThinkingConfig{
				Type:         "enabled",
				BudgetTokens: 5000,
			},
			wantContent:     "Final answer.",
			wantThinking:    "First thought...\nSecond thought...",
			wantTokens:      9, // Approximate
			wantBudgetAlloc: 5000,
			wantBudgetUsed:  9,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			resp, err := processAnthropicResponse(tt.payload, tt.thinkingConfig)
			require.NoError(t, err)
			require.Len(t, resp.Choices, 1)

			choice := resp.Choices[0]
			assert.Equal(t, tt.wantContent, choice.Content)

			if tt.wantThinking != "" {
				thinkingContent, ok := choice.GenerationInfo["ThinkingContent"].(string)
				require.True(t, ok, "ThinkingContent should be present")
				assert.Equal(t, tt.wantThinking, thinkingContent)

				thinkingTokens, ok := choice.GenerationInfo["ThinkingTokens"].(int)
				require.True(t, ok, "ThinkingTokens should be present")
				assert.GreaterOrEqual(t, thinkingTokens, tt.wantTokens-5) // Allow some variance
				assert.LessOrEqual(t, thinkingTokens, tt.wantTokens+5)

				if tt.wantBudgetAlloc > 0 {
					budgetAlloc, ok := choice.GenerationInfo["ThinkingBudgetAllocated"].(int)
					require.True(t, ok, "ThinkingBudgetAllocated should be present")
					assert.Equal(t, tt.wantBudgetAlloc, budgetAlloc)

					budgetUsed, ok := choice.GenerationInfo["ThinkingBudgetUsed"].(int)
					require.True(t, ok, "ThinkingBudgetUsed should be present")
					assert.GreaterOrEqual(t, budgetUsed, tt.wantBudgetUsed-5)
					assert.LessOrEqual(t, budgetUsed, tt.wantBudgetUsed+5)
				}
			} else {
				_, ok := choice.GenerationInfo["ThinkingContent"]
				assert.False(t, ok, "ThinkingContent should not be present")
			}
		})
	}
}

func TestExtractReasoningUsage(t *testing.T) {
	t.Parallel()

	genInfo := map[string]any{
		"ThinkingContent":         "Complex reasoning process...",
		"ThinkingTokens":          500,
		"ThinkingBudgetAllocated": 8000,
		"ThinkingBudgetUsed":      500,
		"OutputTokens":            100,
	}

	usage := llms.ExtractReasoningUsage(genInfo)
	require.NotNil(t, usage)

	assert.Equal(t, 500, usage.ReasoningTokens)
	assert.Equal(t, "Complex reasoning process...", usage.ThinkingContent)
	assert.Equal(t, "Complex reasoning process...", usage.ReasoningContent) // Should be copied
	assert.Equal(t, 8000, usage.BudgetAllocated)
	assert.Equal(t, 500, usage.BudgetUsed)
}

func TestSupportsReasoning(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		model         string
		wantSupported bool
	}{
		{
			name:          "Claude 4 Sonnet",
			model:         "claude-sonnet-4-20250514",
			wantSupported: true,
		},
		{
			name:          "Claude 4 Opus",
			model:         "claude-opus-4-20250101",
			wantSupported: true,
		},
		{
			name:          "Claude 3.7 Sonnet",
			model:         "claude-3-7-sonnet-20250219",
			wantSupported: true,
		},
		{
			name:          "Claude 3.5 Sonnet (old)",
			model:         "claude-3-5-sonnet-20240620",
			wantSupported: false,
		},
		{
			name:          "Claude 3 Haiku",
			model:         "claude-3-haiku-20240307",
			wantSupported: false,
		},
		{
			name:          "Empty model",
			model:         "",
			wantSupported: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			llm := &LLM{model: tt.model}
			assert.Equal(t, tt.wantSupported, llm.SupportsReasoning())
		})
	}
}

// Helper function to create int pointer
func ptrInt(i int) *int {
	return &i
}
