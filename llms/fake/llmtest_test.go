package fake

import (
	"testing"

	"github.com/vxcontrol/langchaingo/testing/llmtest"
)

func TestLLM(t *testing.T) {
	// Fake LLM doesn't need API keys
	// Provide enough responses for all test scenarios
	// Note: llmtest expects "hello" in responses, so we include it in all
	responses := []string{
		"Hello",
		"Hello",
		"Hello",
		"Hello",
		"Hello",
		"Hello",
		"Hello",
		"Hello",
		"Hello",
		"Hello",
		"Hello",
		"Hello",
		"Hello",
		"Hello",
		"Hello",
		"Hello",
		"Hello",
		"Hello",
		"Hello",
		"Hello",
	}
	llm := NewFakeLLM(responses)

	// Test basic functionality only
	llmtest.TestLLM(t, llm)
}
