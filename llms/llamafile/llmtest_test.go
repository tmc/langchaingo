package llamafile

import (
	"testing"

	"github.com/tmc/langchaingo/testing/llmtest"
)

func TestLLM(t *testing.T) {
	// Llamafile uses LLAMAFILE_HOST environment variable for server URL
	// If not set, it defaults to http://127.0.0.1:8080

	llm, err := New()
	if err != nil {
		t.Skipf("Llamafile server not available: %v", err)
	}

	llmtest.TestLLM(t, llm)
}
