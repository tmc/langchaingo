package llmtest_test

import (
	"context"
	"testing"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/testing/llmtest"
)

// Example of basic usage
func ExampleTestLLM() {
	mock := &llmtest.MockLLM{
		CallResponse: "OK",
		GenerateResponse: &llms.ContentResponse{
			Choices: []*llms.ContentChoice{{Content: "Hello"}},
		},
	}

	// This would be called in a test function with testing.T
	_ = mock
}

// Example of using MockLLM with custom functions
func ExampleMockLLM_customFunction() {
	mock := &llmtest.MockLLM{
		CallFunc: func(ctx context.Context, prompt string, options ...llms.CallOption) (string, error) {
			// Custom logic based on prompt
			if prompt == "hello" {
				return "Hi there!", nil
			}
			return "OK", nil
		},
	}

	// This would be used in tests
	_ = mock
}

// Example of MockLLM with feature flags
func ExampleMockLLM_features() {
	mock := &llmtest.MockLLM{
		CallResponse:             "OK",
		SupportsToolCalls:        true,
		SupportsReasoningMode:    true,
		SupportsStructuredOutput: true,
		SupportsMultimodalInput:  true,
	}

	// This would be used in tests
	_ = mock
}

// Example of benchmarking
func ExampleBenchmarkLLM() {
	// In a real benchmark function:
	// func BenchmarkMyLLM(b *testing.B) {
	//     llm := &llmtest.MockLLM{CallResponse: "OK"}
	//     llmtest.BenchmarkLLM(b, llm)
	// }
}

// Example of benchmarking with options
func ExampleBenchmarkLLMWithOptions() {
	// In a real benchmark function:
	// func BenchmarkMyLLM(b *testing.B) {
	//     llm := &llmtest.MockLLM{CallResponse: "OK"}
	//     llmtest.BenchmarkLLMWithOptions(b, llm, llmtest.BenchmarkOptions{
	//         Prompt: "Custom benchmark prompt",
	//         MaxTokens: 100,
	//     })
	// }
}

// Example of ValidateLLM
func ExampleValidateLLM() {
	mock := &llmtest.MockLLM{
		CallResponse: "OK",
		GenerateResponse: &llms.ContentResponse{
			Choices: []*llms.ContentChoice{{Content: "Hello"}},
		},
	}

	err := llmtest.ValidateLLM(mock)
	if err != nil {
		// Handle validation error
		_ = err
	}
}

// Example test showing all new features
func TestNewFeatures(t *testing.T) {
	t.Run("StructuredOutput", func(t *testing.T) {
		mock := &llmtest.MockLLM{
			GenerateResponse: &llms.ContentResponse{
				Choices: []*llms.ContentChoice{{
					Content: `{"name": "test", "value": 42}`,
				}},
			},
			SupportsStructuredOutput: true,
		}

		// Test that MockLLM tracks calls
		ctx := context.Background()
		_, _ = mock.GenerateContent(ctx, []llms.MessageContent{
			{
				Role:  llms.ChatMessageTypeHuman,
				Parts: []llms.ContentPart{llms.TextPart("test")},
			},
		})

		if mock.GenerateCount != 1 {
			t.Errorf("Expected 1 call, got %d", mock.GenerateCount)
		}
	})

	t.Run("Multimodal", func(t *testing.T) {
		mock := &llmtest.MockLLM{
			GenerateResponse: &llms.ContentResponse{
				Choices: []*llms.ContentChoice{{
					Content: "I see a red image",
				}},
			},
			SupportsMultimodalInput: true,
		}

		ctx := context.Background()
		_, _ = mock.GenerateContent(ctx, []llms.MessageContent{
			{
				Role: llms.ChatMessageTypeHuman,
				Parts: []llms.ContentPart{
					llms.ImageURLPart("data:image/png;base64,iVBORw0KGgo="),
					llms.TextPart("What do you see?"),
				},
			},
		})

		if mock.GenerateCount != 1 {
			t.Errorf("Expected 1 call, got %d", mock.GenerateCount)
		}
	})

	t.Run("Reasoning", func(t *testing.T) {
		mock := &llmtest.MockLLM{
			GenerateResponse: &llms.ContentResponse{
				Choices: []*llms.ContentChoice{{
					Content: "42",
					GenerationInfo: map[string]any{
						"ThinkingTokens": 100,
					},
				}},
			},
			SupportsReasoningMode: true,
		}

		if !mock.SupportsReasoning() {
			t.Error("Expected reasoning support")
		}
	})
}
