package ollama

import (
	"os"
	"testing"

	"github.com/tmc/langchaingo/testing/llmtest"
)

func TestLLM(t *testing.T) {
	serverURL := os.Getenv("OLLAMA_HOST")
	if serverURL == "" {
		serverURL = "http://localhost:11434"
	}

	llm, err := New(
		WithServerURL(serverURL),
		WithModel("gpt-oss:20b"),
	)
	if err != nil {
		t.Skipf("Ollama not available: %v", err)
	}

	llmtest.TestLLM(t, llm)
}
