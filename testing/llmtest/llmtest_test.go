package llmtest

import (
	"os"
	"testing"

	"github.com/vendasta/langchaingo/llms"
)

// TestMockLLM tests the mock implementation.
func TestMockLLM(t *testing.T) {
	mock := &MockLLM{
		CallResponse: "OK",
		GenerateResponse: &llms.ContentResponse{
			Choices: []*llms.ContentChoice{
				{
					Content: "Hello",
					GenerationInfo: map[string]interface{}{
						"TotalTokens": 10,
					},
				},
			},
		},
	}

	TestLLM(t, mock)
}

// TestValidateLLM tests the validation function.
func TestValidateLLM(t *testing.T) {
	// Test with nil model
	if err := ValidateLLM(nil); err == nil {
		t.Error("ValidateLLM should fail with nil model")
	}

	// Test with valid mock
	mock := &MockLLM{
		CallResponse: "OK",
		GenerateResponse: &llms.ContentResponse{
			Choices: []*llms.ContentChoice{
				{
					Content: "response",
				},
			},
		},
	}

	if err := ValidateLLM(mock); err != nil {
		t.Errorf("ValidateLLM failed with valid mock: %v", err)
	}
}

// Integration tests with real providers (require API keys)

func TestAnthropicIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	if os.Getenv("ANTHROPIC_API_KEY") == "" {
		t.Skip("ANTHROPIC_API_KEY not set")
	}

	// Import is handled in the actual test files for each provider
}

func TestOpenAIIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	if os.Getenv("OPENAI_API_KEY") == "" {
		t.Skip("OPENAI_API_KEY not set")
	}

	// Import is handled in the actual test files for each provider
}
