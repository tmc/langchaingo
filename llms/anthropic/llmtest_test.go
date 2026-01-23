package anthropic

import (
	"os"
	"testing"

	"github.com/vxcontrol/langchaingo/testing/llmtest"
)

func TestLLM(t *testing.T) {
	if os.Getenv("ANTHROPIC_API_KEY") == "" {
		t.Skip("ANTHROPIC_API_KEY not set")
	}

	llm, err := New(WithModel("claude-haiku-4-5"))
	if err != nil {
		t.Fatalf("Failed to create Anthropic LLM: %v", err)
	}

	// Test with automatic capability discovery
	llmtest.TestLLM(t, llm)
}
