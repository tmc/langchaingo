package openai

import (
	"encoding/json"
	"testing"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai/internal/openaiclient"
)

// Helper function to create and test token options mapping
func testTokenOptionsMapping(t *testing.T, options []llms.CallOption, expectedMaxCompletionTokens, expectedMaxTokens int, description string) {
	// Apply options to get CallOptions
	opts := llms.CallOptions{}
	for _, opt := range options {
		opt(&opts)
	}

	// Create the request struct using the same logic as the OpenAI client
	req := createTestChatRequest(opts)

	// Verify the field values
	if req.MaxCompletionTokens != expectedMaxCompletionTokens {
		t.Errorf("MaxCompletionTokens = %v, want %v. %s",
			req.MaxCompletionTokens, expectedMaxCompletionTokens, description)
	}

	// Use a helper to avoid directly accessing deprecated field
	actualMaxTokens := getMaxTokensForTest(req)
	if actualMaxTokens != expectedMaxTokens {
		t.Errorf("MaxTokens = %v, want %v. %s",
			actualMaxTokens, expectedMaxTokens, description)
	}

	// Verify that both fields are never set simultaneously
	if req.MaxCompletionTokens > 0 && actualMaxTokens > 0 {
		t.Errorf("Both MaxCompletionTokens (%v) and MaxTokens (%v) are set, which would cause API error",
			req.MaxCompletionTokens, actualMaxTokens)
	}
}

// Helper function to create a chat request for testing
func createTestChatRequest(opts llms.CallOptions) *openaiclient.ChatRequest {
	return &openaiclient.ChatRequest{
		Model:    "gpt-3.5-turbo",
		Messages: []*openaiclient.ChatMessage{{Role: "user", Content: "test"}},
		MaxCompletionTokens: func() int {
			// Check if this should explicitly use max_completion_tokens field
			if opts.Metadata != nil {
				if useMaxCompletion, ok := opts.Metadata["_openai_use_max_completion_tokens"].(bool); ok && useMaxCompletion {
					return opts.MaxTokens
				}
			}
			// Default for backward compatibility: use max_completion_tokens if no deprecated flag is set
			if opts.Metadata == nil || opts.Metadata["_openai_use_deprecated_max_tokens"] == nil {
				if opts.MaxTokens > 0 {
					return opts.MaxTokens
				}
			}
			return 0
		}(),
		MaxTokens: func() int {
			// Only use the deprecated max_tokens field when explicitly requested
			if opts.Metadata != nil {
				if useDeprecated, ok := opts.Metadata["_openai_use_deprecated_max_tokens"].(bool); ok && useDeprecated {
					return opts.MaxTokens
				}
			}
			return 0
		}(),
	}
}

// Helper function to get MaxTokens without directly accessing deprecated field
func getMaxTokensForTest(req *openaiclient.ChatRequest) int {
	// Use JSON marshaling to avoid direct access to deprecated field
	data, _ := json.Marshal(req)
	var result map[string]interface{}
	json.Unmarshal(data, &result) //nolint:errcheck // Test helper
	if val, ok := result["max_tokens"]; ok {
		return int(val.(float64))
	}
	return 0
}

// Helper function to test JSON serialization
func testJSONSerialization(t *testing.T, options []llms.CallOption, expectedMaxCompletionInJSON, expectedMaxTokensInJSON bool, expectedValues map[string]int) {
	// Apply options
	opts := llms.CallOptions{}
	for _, opt := range options {
		opt(&opts)
	}

	// Create request using same logic as client
	req := createTestChatRequest(opts)

	// Marshal to JSON
	jsonData, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal ChatRequest: %v", err)
	}

	// Parse JSON back to map to check field presence
	var jsonMap map[string]interface{}
	if err := json.Unmarshal(jsonData, &jsonMap); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	// Check max_completion_tokens field
	if expectedMaxCompletionInJSON {
		value, exists := jsonMap["max_completion_tokens"]
		if !exists {
			t.Error("Expected max_completion_tokens field in JSON, but it was not present")
		} else if int(value.(float64)) != expectedValues["max_completion_tokens"] {
			t.Errorf("max_completion_tokens = %v, want %v", value, expectedValues["max_completion_tokens"])
		}
	} else {
		if _, exists := jsonMap["max_completion_tokens"]; exists {
			t.Error("max_completion_tokens field should not be present in JSON")
		}
	}

	// Check max_tokens field
	if expectedMaxTokensInJSON {
		value, exists := jsonMap["max_tokens"]
		if !exists {
			t.Error("Expected max_tokens field in JSON, but it was not present")
		} else if int(value.(float64)) != expectedValues["max_tokens"] {
			t.Errorf("max_tokens = %v, want %v", value, expectedValues["max_tokens"])
		}
	} else {
		if _, exists := jsonMap["max_tokens"]; exists {
			t.Error("max_tokens field should not be present in JSON")
		}
	}
}

func TestOpenAITokenOptionsMapping(t *testing.T) {
	tests := []struct {
		name                        string
		options                     []llms.CallOption
		expectedMaxCompletionTokens int
		expectedMaxTokens           int
		description                 string
	}{
		{
			name:                        "WithMaxCompletionTokens only",
			options:                     []llms.CallOption{WithMaxCompletionTokens(150)},
			expectedMaxCompletionTokens: 150,
			expectedMaxTokens:           0,
			description:                 "Should set max_completion_tokens and not max_tokens",
		},
		{
			name:                        "WithMaxTokens only (backward compatibility)",
			options:                     []llms.CallOption{llms.WithMaxTokens(100)},
			expectedMaxCompletionTokens: 100,
			expectedMaxTokens:           0,
			description:                 "Should set max_completion_tokens for backward compatibility (recommended field)",
		},
		{
			name:                        "WithDeprecatedMaxTokens only",
			options:                     []llms.CallOption{WithDeprecatedMaxTokens(200)},
			expectedMaxCompletionTokens: 0,
			expectedMaxTokens:           200,
			description:                 "Should set max_tokens when explicitly using deprecated field",
		},
		{
			name: "Generic WithMaxTokens with OpenAI WithMaxCompletionTokens",
			options: []llms.CallOption{
				llms.WithMaxTokens(100),
				WithMaxCompletionTokens(150),
			},
			expectedMaxCompletionTokens: 150,
			expectedMaxTokens:           0,
			description:                 "OpenAI-specific option should override generic option",
		},
		{
			name:                        "Neither option set",
			options:                     []llms.CallOption{},
			expectedMaxCompletionTokens: 0,
			expectedMaxTokens:           0,
			description:                 "Should not set either field when no token options provided",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testTokenOptionsMapping(t, tt.options, tt.expectedMaxCompletionTokens, tt.expectedMaxTokens, tt.description)
		})
	}
}

func TestOpenAITokenOptionsJSONSerialization(t *testing.T) {
	tests := []struct {
		name                        string
		options                     []llms.CallOption
		expectedMaxCompletionInJSON bool
		expectedMaxTokensInJSON     bool
		expectedValues              map[string]int
	}{
		{
			name:                        "WithMaxCompletionTokens serialization",
			options:                     []llms.CallOption{WithMaxCompletionTokens(150)},
			expectedMaxCompletionInJSON: true,
			expectedMaxTokensInJSON:     false,
			expectedValues:              map[string]int{"max_completion_tokens": 150},
		},
		{
			name:                        "WithDeprecatedMaxTokens serialization",
			options:                     []llms.CallOption{WithDeprecatedMaxTokens(100)},
			expectedMaxCompletionInJSON: false,
			expectedMaxTokensInJSON:     true,
			expectedValues:              map[string]int{"max_tokens": 100},
		},
		{
			name:                        "Generic WithMaxTokens serialization",
			options:                     []llms.CallOption{llms.WithMaxTokens(100)},
			expectedMaxCompletionInJSON: true,
			expectedMaxTokensInJSON:     false,
			expectedValues:              map[string]int{"max_completion_tokens": 100},
		},
		{
			name:                        "No token options",
			options:                     []llms.CallOption{},
			expectedMaxCompletionInJSON: false,
			expectedMaxTokensInJSON:     false,
			expectedValues:              map[string]int{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testJSONSerialization(t, tt.options, tt.expectedMaxCompletionInJSON, tt.expectedMaxTokensInJSON, tt.expectedValues)
		})
	}
}

func TestBackwardCompatibilityWithNewArchitecture(t *testing.T) {
	t.Run("Generic llms.WithMaxTokens still works", func(t *testing.T) {
		opts := llms.CallOptions{}
		llms.WithMaxTokens(100)(&opts)

		// Test that it maps to max_completion_tokens for backward compatibility
		maxCompletionTokens := func() int {
			// Check if this should explicitly use max_completion_tokens field
			if opts.Metadata != nil {
				if useMaxCompletion, ok := opts.Metadata["_openai_use_max_completion_tokens"].(bool); ok && useMaxCompletion {
					return opts.MaxTokens
				}
			}
			// Default for backward compatibility: use max_completion_tokens if no deprecated flag is set
			if opts.Metadata == nil || opts.Metadata["_openai_use_deprecated_max_tokens"] == nil {
				if opts.MaxTokens > 0 {
					return opts.MaxTokens
				}
			}
			return 0
		}()

		if maxCompletionTokens != 100 {
			t.Errorf("Expected MaxCompletionTokens = 100 for backward compatibility, got %v", maxCompletionTokens)
		}
	})

	t.Run("OpenAI-specific options work correctly", func(t *testing.T) {
		opts := llms.CallOptions{}
		WithMaxCompletionTokens(150)(&opts)

		// Test that it maps to max_completion_tokens
		maxCompletionTokens := func() int {
			if opts.Metadata != nil {
				if useMaxCompletion, ok := opts.Metadata["_openai_use_max_completion_tokens"].(bool); ok && useMaxCompletion {
					return opts.MaxTokens
				}
			}
			return 0
		}()

		if maxCompletionTokens != 150 {
			t.Errorf("Expected MaxCompletionTokens = 150, got %v", maxCompletionTokens)
		}
	})

	t.Run("Deprecated max_tokens option works correctly", func(t *testing.T) {
		opts := llms.CallOptions{}
		WithDeprecatedMaxTokens(200)(&opts)

		// Test that it maps to max_tokens
		maxTokens := func() int {
			if opts.Metadata != nil {
				if useDeprecated, ok := opts.Metadata["_openai_use_deprecated_max_tokens"].(bool); ok && useDeprecated {
					return opts.MaxTokens
				}
			}
			return 0
		}()

		if maxTokens != 200 {
			t.Errorf("Expected MaxTokens = 200, got %v", maxTokens)
		}
	})
}
