package anthropic

import (
	"os"
	"testing"

	"github.com/vendasta/langchaingo/testing/llmtest"
)

func TestLLM(t *testing.T) {
	if os.Getenv("ANTHROPIC_API_KEY") == "" {
		t.Skip("ANTHROPIC_API_KEY not set")
	}

	llm, err := New(WithModel("claude-3-haiku-20240307"))
	if err != nil {
		t.Fatalf("Failed to create Anthropic LLM: %v", err)
	}

	// Test with automatic capability discovery
	llmtest.TestLLM(t, llm)
}
