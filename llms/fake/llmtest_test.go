package fake

import (
	"testing"

	"github.com/tmc/langchaingo/testing/llmtest"
)

func TestLLM(t *testing.T) {
	// Fake LLM doesn't need API keys
	// Configure with enough responses for all the tests
	// Note: Since tests run in parallel and responses cycle, we provide
	// enough generic responses that will work for most tests
	responses := []string{
		"OK",
		"Hello",
		"test",
		"1 2 3",
		"42",
		"{ \"test\": true }",
		"red",
	}
	llm := NewFakeLLM(responses)

	// Test basic functionality
	llmtest.TestLLM(t, llm)
}
