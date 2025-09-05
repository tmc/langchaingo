package ernie

import (
	"os"
	"testing"

	"github.com/tmc/langchaingo/testing/llmtest"
)

func TestLLM(t *testing.T) {
	apiKey := os.Getenv("ERNIE_API_KEY")
	if apiKey == "" {
		t.Skip("ERNIE_API_KEY not set")
	}

	secretKey := os.Getenv("ERNIE_SECRET_KEY")
	if secretKey == "" {
		t.Skip("ERNIE_SECRET_KEY not set")
	}

	llm, err := New(WithAKSK(apiKey, secretKey))
	if err != nil {
		t.Fatalf("Failed to create Ernie LLM: %v", err)
	}

	llmtest.TestLLM(t, llm)
}
