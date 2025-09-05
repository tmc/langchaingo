package cohere

import (
	"os"
	"testing"

	"github.com/tmc/langchaingo/testing/llmtest"
)

func TestLLM(t *testing.T) {
	if os.Getenv("COHERE_API_KEY") == "" {
		t.Skip("COHERE_API_KEY not set")
	}

	llm, err := New()
	if err != nil {
		t.Fatalf("Failed to create Cohere LLM: %v", err)
	}

	llmtest.TestLLM(t, llm)
}
