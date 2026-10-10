package ollama

import (
	"context"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/tmc/langchaingo/testing/llmtest"
)

// liveServerURL returns the Ollama server URL for live tests.
// It skips the test if no server answers there.
func liveServerURL(t *testing.T) string {
	t.Helper()
	serverURL := os.Getenv("OLLAMA_HOST")
	if serverURL == "" {
		serverURL = "http://localhost:11434"
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, serverURL+"/api/version", nil)
	if err != nil {
		t.Skipf("Ollama not available: %v", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Skipf("Ollama not available: %v", err)
	}
	resp.Body.Close()
	return serverURL
}

func TestLLM(t *testing.T) {
	llm, err := New(
		WithServerURL(liveServerURL(t)),
		WithModel("gpt-oss:20b"),
	)
	if err != nil {
		t.Skipf("Ollama not available: %v", err)
	}

	llmtest.TestLLM(t, llm)
}
