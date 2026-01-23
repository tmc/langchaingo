package ollama

import (
	"net/http"
	"net/url"
	"os"
	"testing"

	"github.com/ollama/ollama/api"
	"github.com/vxcontrol/langchaingo/testing/llmtest"
)

func checkIfModelExists(t *testing.T, model string, serverURL string) bool {
	t.Helper()

	url, err := url.Parse(serverURL)
	if err != nil {
		t.Fatalf("Failed to parse server URL: %v", err)
	}

	client := api.NewClient(url, http.DefaultClient)

	if _, err := client.Show(t.Context(), &api.ShowRequest{Model: model}); err == nil {
		return true
	}

	return false
}

func TestLLM(t *testing.T) {
	serverURL := os.Getenv("OLLAMA_HOST")
	if serverURL == "" {
		serverURL = "http://localhost:11434"
	}

	model := "gpt-oss:20b"

	if !checkIfModelExists(t, model, serverURL) {
		t.Skipf("Model %s not available", model)
	}

	llm, err := New(
		WithServerURL(serverURL),
		WithModel(model),
	)
	if err != nil {
		t.Fatalf("Failed to create LLM: %v", err)
	}

	llmtest.TestLLM(t, llm)
}
