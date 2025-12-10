package googleai

import (
	"context"
	"os"
	"testing"

	"github.com/vendasta/langchaingo/testing/llmtest"
)

func TestLLM(t *testing.T) {
	if os.Getenv("GOOGLE_API_KEY") == "" {
		t.Skip("GOOGLE_API_KEY not set")
	}

	ctx := context.Background()
	llm, err := New(ctx,
		WithAPIKey(os.Getenv("GOOGLE_API_KEY")),
		WithDefaultModel("gemini-3-pro-preview"),
	)
	if err != nil {
		t.Fatalf("Failed to create Google AI LLM: %v", err)
	}
	defer llm.Close()

	llmtest.TestLLM(t, llm)
}
