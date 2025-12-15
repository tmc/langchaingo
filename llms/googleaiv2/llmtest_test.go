package googleaiv2

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
	// Use gemini-2.0-flash which is the default and known to work
	// googleaiv2 uses the new SDK and can work with both gemini-2.x and gemini-3.x models
	llm, err := New(ctx,
		WithAPIKey(os.Getenv("GOOGLE_API_KEY")),
		WithDefaultModel("gemini-2.0-flash"),
	)
	if err != nil {
		t.Fatalf("Failed to create Google AI LLM: %v", err)
	}
	defer llm.Close()

	llmtest.TestLLM(t, llm)
}
