package llamafile

import (
	"os"
	"testing"

	"github.com/vxcontrol/langchaingo/testing/llmtest"
)

func TestLLM(t *testing.T) {
	if os.Getenv("LLAMAFILE_HOST") == "" {
		t.Skip("LLAMAFILE_HOST not set")
	}

	llm, err := New()
	if err != nil {
		t.Skipf("Llamafile server not available: %v", err)
	}

	llmtest.TestLLM(t, llm)
}
