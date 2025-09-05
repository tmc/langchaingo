package maritaca

import (
	"os"
	"testing"

	"github.com/tmc/langchaingo/testing/llmtest"
)

func TestLLM(t *testing.T) {
	if os.Getenv("MARITACA_API_KEY") == "" {
		t.Skip("MARITACA_API_KEY not set")
	}

	llm, err := New(
		WithToken(os.Getenv("MARITACA_API_KEY")),
		WithModel("sabia-3"),
	)
	if err != nil {
		t.Fatalf("Failed to create Maritaca LLM: %v", err)
	}

	llmtest.TestLLM(t, llm)
}
