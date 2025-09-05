package cloudflare

import (
	"os"
	"testing"

	"github.com/tmc/langchaingo/testing/llmtest"
)

func TestLLM(t *testing.T) {
	apiToken := os.Getenv("CLOUDFLARE_API_TOKEN")
	if apiToken == "" {
		t.Skip("CLOUDFLARE_API_TOKEN not set")
	}

	accountID := os.Getenv("CLOUDFLARE_ACCOUNT_ID")
	if accountID == "" {
		t.Skip("CLOUDFLARE_ACCOUNT_ID not set")
	}

	llm, err := New(WithToken(apiToken), WithAccountID(accountID))
	if err != nil {
		t.Fatalf("Failed to create Cloudflare LLM: %v", err)
	}

	llmtest.TestLLM(t, llm)
}