package mistral

import (
	"os"
	"testing"

	"github.com/vendasta/langchaingo/testing/llmtest"
)

func TestLLM(t *testing.T) {
	if os.Getenv("MISTRAL_API_KEY") == "" {
		t.Skip("MISTRAL_API_KEY not set")
	}

	llm, err := New(
		WithAPIKey(os.Getenv("MISTRAL_API_KEY")),
	)
	if err != nil {
		t.Fatalf("Failed to create Mistral LLM: %v", err)
	}

	llmtest.TestLLM(t, llm)
}
