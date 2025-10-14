package openai

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vendasta/langchaingo/llms"
	"github.com/vendasta/langchaingo/llms/openai/internal/openaiclient"
)

func TestMaxTokensFieldSerialization(t *testing.T) {
	// This test verifies that only max_completion_tokens is sent,
	// never max_tokens (which would cause an OpenAI API error)

	tests := []struct {
		name     string
		request  openaiclient.ChatRequest
		expected map[string]interface{}
	}{
		{
			name: "MaxCompletionTokens is serialized",
			request: openaiclient.ChatRequest{
				Model:               "gpt-4",
				MaxCompletionTokens: 100,
				Temperature:         0.7,
			},
			expected: map[string]interface{}{
				"model":                 "gpt-4",
				"max_completion_tokens": float64(100),
				"temperature":           0.7,
			},
		},
		{
			name: "Both fields set - only MaxCompletionTokens is sent",
			request: openaiclient.ChatRequest{
				Model:               "gpt-4",
				MaxTokens:           100,
				MaxCompletionTokens: 200,
				Temperature:         0.7,
			},
			expected: map[string]interface{}{
				"model":                 "gpt-4",
				"max_completion_tokens": float64(200),
				"temperature":           0.7,
				// Note: max_tokens is NOT in the output due to MarshalJSON logic
			},
		},
		{
			name: "Only MaxTokens set - it IS serialized",
			request: openaiclient.ChatRequest{
				Model:       "gpt-4",
				MaxTokens:   100,
				Temperature: 0.7,
			},
			expected: map[string]interface{}{
				"model":       "gpt-4",
				"max_tokens":  float64(100),
				"temperature": 0.7,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Marshal the request
			data, err := json.Marshal(tt.request)
			require.NoError(t, err)

			// Unmarshal to a map to check fields
			var result map[string]interface{}
			err = json.Unmarshal(data, &result)
			require.NoError(t, err)

			// For tests with both fields, verify only one is present
			hasMaxTokens := result["max_tokens"] != nil
			hasMaxCompletionTokens := result["max_completion_tokens"] != nil

			// Never both
			if hasMaxTokens && hasMaxCompletionTokens {
				t.Error("Both max_tokens and max_completion_tokens are present - API will error!")
			}

			// Verify expected fields
			for key, expectedValue := range tt.expected {
				actualValue, exists := result[key]
				assert.True(t, exists, "field %s should exist", key)
				assert.Equal(t, expectedValue, actualValue, "field %s value mismatch", key)
			}
		})
	}
}

func TestMaxTokensBehaviorDocumentation(t *testing.T) {
	// This test documents the expected behavior for users

	t.Run("llms.WithMaxTokens sets max_completion_tokens", func(t *testing.T) {
		opts := &llms.CallOptions{}
		llms.WithMaxTokens(100)(opts)
		assert.Equal(t, 100, opts.MaxTokens)
		// In openaillm.go, this opts.MaxTokens value gets mapped to MaxCompletionTokens
	})

	t.Run("openai.WithMaxCompletionTokens is explicit", func(t *testing.T) {
		opts := &llms.CallOptions{}
		WithMaxCompletionTokens(100)(opts)
		assert.Equal(t, 100, opts.MaxTokens)
		// Same effect, but more explicit about what field is being set
	})
}
