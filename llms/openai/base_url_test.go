package openai

import (
	"testing"
)

func TestIsUsingCustomBaseURL(t *testing.T) {
	tests := []struct {
		name     string
		options  []Option
		expected bool
	}{
		{
			name:     "Default OpenAI URL",
			options:  []Option{},
			expected: false,
		},
		{
			name:     "Explicit default OpenAI URL",
			options:  []Option{WithBaseURL("https://api.openai.com/v1")},
			expected: false,
		},
		{
			name:     "Custom base URL (OpenRouter)",
			options:  []Option{WithBaseURL("https://openrouter.ai/api/v1")},
			expected: true,
		},
		{
			name:     "Custom base URL (local)",
			options:  []Option{WithBaseURL("http://localhost:8080/v1")},
			expected: true,
		},
		{
			name:     "Azure URL",
			options:  []Option{WithBaseURL("https://myinstance.openai.azure.com")},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create client with test options
			opts := append(tt.options, WithToken("test-token"))
			llm, err := New(opts...)
			if err != nil {
				t.Fatalf("Failed to create OpenAI client: %v", err)
			}

			result := llm.IsUsingCustomBaseURL()
			if result != tt.expected {
				t.Errorf("IsUsingCustomBaseURL() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGetBaseURL(t *testing.T) {
	tests := []struct {
		name        string
		options     []Option
		expectedURL string
	}{
		{
			name:        "Default OpenAI URL",
			options:     []Option{},
			expectedURL: "https://api.openai.com/v1",
		},
		{
			name:        "Custom base URL (OpenRouter)",
			options:     []Option{WithBaseURL("https://openrouter.ai/api/v1")},
			expectedURL: "https://openrouter.ai/api/v1",
		},
		{
			name:        "Custom base URL (local)",
			options:     []Option{WithBaseURL("http://localhost:8080/v1")},
			expectedURL: "http://localhost:8080/v1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create client with test options
			opts := append(tt.options, WithToken("test-token"))
			llm, err := New(opts...)
			if err != nil {
				t.Fatalf("Failed to create OpenAI client: %v", err)
			}

			result := llm.GetBaseURL()
			if result != tt.expectedURL {
				t.Errorf("GetBaseURL() = %v, want %v", result, tt.expectedURL)
			}
		})
	}
}
