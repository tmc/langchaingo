package fake

import (
	"context"
	"testing"

	"github.com/tmc/langchaingo/testing/llmtest"
)

func TestLLM(t *testing.T) {
	llm := NewFakeLLM([]string{"OK", "Grace", "1 2 3"})
	if err := llmtest.TestModel(context.Background(), llm, "multiturn", "streaming"); err != nil {
		t.Fatal(err)
	}
}
