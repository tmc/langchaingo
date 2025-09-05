package local

import (
	"os"
	"testing"

	"github.com/tmc/langchaingo/testing/llmtest"
)

func TestLLM(t *testing.T) {
	binPath := os.Getenv("LOCAL_LLM_BIN")
	if binPath == "" {
		t.Skip("LOCAL_LLM_BIN not set")
	}

	llm, err := New(WithBin(binPath))
	if err != nil {
		t.Fatalf("Failed to create Local LLM: %v", err)
	}

	llmtest.TestLLM(t, llm)
}