package huggingface

import (
	"os"
	"testing"

	"github.com/vendasta/langchaingo/testing/llmtest"
)

func TestLLM(t *testing.T) {
	if os.Getenv("HUGGINGFACEHUB_API_TOKEN") == "" {
		t.Skip("HUGGINGFACEHUB_API_TOKEN not set")
	}

	llm, err := New(WithModel("gpt2"))
	if err != nil {
		t.Fatalf("Failed to create HuggingFace LLM: %v", err)
	}

	llmtest.TestLLM(t, llm)
}
