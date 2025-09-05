package openai

import (
	"os"
	"testing"

	"github.com/tmc/langchaingo/testing/llmtest"
)

func TestLLM(t *testing.T) {
	if os.Getenv("OPENAI_API_KEY") == "" {
		t.Skip("OPENAI_API_KEY not set")
	}

	llm, err := New(WithModel("gpt-3.5-turbo"))
	if err != nil {
		t.Fatalf("Failed to create OpenAI LLM: %v", err)
	}

	llmtest.TestLLM(t, llm)
}
