package openai

import (
	"context"
	"testing"

	"github.com/tmc/langchaingo/llms"
)

// Mock LLM for testing
type mockNonOpenAILLM struct{}

func (m *mockNonOpenAILLM) Call(ctx context.Context, prompt string, options ...llms.CallOption) (string, error) {
	return "", nil
}

func (m *mockNonOpenAILLM) GenerateContent(ctx context.Context, messages []llms.MessageContent, options ...llms.CallOption) (*llms.ContentResponse, error) {
	return nil, nil
}

func TestIsOpenAIModel(t *testing.T) {
	// Test with OpenAI model
	openaiLLM, err := New(WithToken("test-token"))
	if err != nil {
		t.Fatalf("Failed to create OpenAI LLM: %v", err)
	}

	if !IsOpenAIModel(openaiLLM) {
		t.Error("IsOpenAIModel should return true for OpenAI LLM")
	}

	// Test with non-OpenAI model
	mockLLM := &mockNonOpenAILLM{}
	if IsOpenAIModel(mockLLM) {
		t.Error("IsOpenAIModel should return false for non-OpenAI LLM")
	}
}

func TestGetOpenAIBaseURL(t *testing.T) {
	tests := []struct {
		name        string
		options     []Option
		expectedURL string
	}{
		{
			name:        "Default OpenAI URL",
			options:     []Option{WithToken("test-token")},
			expectedURL: "https://api.openai.com/v1",
		},
		{
			name:        "Custom base URL",
			options:     []Option{WithToken("test-token"), WithBaseURL("https://openrouter.ai/api/v1")},
			expectedURL: "https://openrouter.ai/api/v1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			openaiLLM, err := New(tt.options...)
			if err != nil {
				t.Fatalf("Failed to create OpenAI LLM: %v", err)
			}

			url := GetOpenAIBaseURL(openaiLLM)
			if url != tt.expectedURL {
				t.Errorf("GetOpenAIBaseURL() = %v, want %v", url, tt.expectedURL)
			}
		})
	}

	// Test with non-OpenAI model
	mockLLM := &mockNonOpenAILLM{}
	url := GetOpenAIBaseURL(mockLLM)
	if url != "" {
		t.Errorf("GetOpenAIBaseURL should return empty string for non-OpenAI LLM, got %v", url)
	}
}

func TestIsOpenAIUsingCustomBaseURL(t *testing.T) {
	tests := []struct {
		name     string
		options  []Option
		expected bool
	}{
		{
			name:     "Default OpenAI URL",
			options:  []Option{WithToken("test-token")},
			expected: false,
		},
		{
			name:     "Custom base URL",
			options:  []Option{WithToken("test-token"), WithBaseURL("https://openrouter.ai/api/v1")},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			openaiLLM, err := New(tt.options...)
			if err != nil {
				t.Fatalf("Failed to create OpenAI LLM: %v", err)
			}

			result := IsOpenAIUsingCustomBaseURL(openaiLLM)
			if result != tt.expected {
				t.Errorf("IsOpenAIUsingCustomBaseURL() = %v, want %v", result, tt.expected)
			}
		})
	}

	// Test with non-OpenAI model
	mockLLM := &mockNonOpenAILLM{}
	result := IsOpenAIUsingCustomBaseURL(mockLLM)
	if result {
		t.Error("IsOpenAIUsingCustomBaseURL should return false for non-OpenAI LLM")
	}
}
