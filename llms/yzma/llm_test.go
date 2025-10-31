package yzma

import (
	"os"
	"testing"

	"github.com/tmc/langchaingo/testing/llmtest"
)

func TestLLM(t *testing.T) {
	testModel := os.Getenv("YZMA_TEST_MODEL")
	if testModel == "" {
		t.Skip("YZMA_TEST_MODEL not set to point to test model")
	}

	llm, err := New(WithModel(testModel))
	if err != nil {
		t.Fatalf("Failed to create yzma LLM: %v", err)
	}
	defer llm.Close()

	llmtest.TestLLM(t, llm)
}
