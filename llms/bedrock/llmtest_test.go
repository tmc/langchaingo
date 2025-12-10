package bedrock

import (
	"os"
	"testing"

	"github.com/tmc/langchaingo/testing/llmtest"
)

func TestLLM(t *testing.T) {
	if os.Getenv("AWS_REGION") == "" {
		t.Skip("AWS_REGION not set")
	}

	llm, err := New(
		WithModel("anthropic.claude-3-haiku-20240307-v1:0"),
	)
	if err != nil {
		t.Fatalf("Failed to create Bedrock LLM: %v", err)
	}

	llmtest.TestLLM(t, llm)
}
