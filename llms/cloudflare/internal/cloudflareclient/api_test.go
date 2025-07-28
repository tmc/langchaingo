package cloudflareclient

import (
	"context"
	"net/http"
	"os"
	"testing"

	"github.com/0xDezzy/langchaingo/internal/httprr"
)

func TestClient_GenerateContent(t *testing.T) {
	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "CLOUDFLARE_API_KEY")

	rr := httprr.OpenForTest(t, http.DefaultTransport)
	defer rr.Close()

	if !rr.Recording() {
		t.Parallel()
	}

	ctx := context.Background()

	// Use test credentials when not recording
	accountID := "test-account-id"
	token := "test-api-token"
	baseURL := "https://api.cloudflare.com/client/v4/accounts"
	modelName := "test-model"

	// Use real credentials when recording
	if rr.Recording() {
		if id := os.Getenv("CLOUDFLARE_ACCOUNT_ID"); id != "" {
			accountID = id
		}
		if tok := os.Getenv("CLOUDFLARE_API_KEY"); tok != "" {
			token = tok
		}
		if model := os.Getenv("CLOUDFLARE_MODEL_NAME"); model != "" {
			modelName = model
		}
	}

	client := NewClient(rr.Client(), accountID, baseURL, token, modelName, "")

	t.Run("test generate content success", func(t *testing.T) {
		request := &GenerateContentRequest{
			Messages: []Message{
				{Role: "system", Content: "You are a helpful assistant."},
				{Role: "user", Content: "Hello, how are you?"},
			},
		}

		response, err := client.GenerateContent(ctx, request)
		if err != nil {
			t.Fatalf("GenerateContent() error = %v", err)
		}

		if response == nil {
			t.Fatal("GenerateContent() response is nil")
		}

		if response.Result.Response == "" {
			t.Error("GenerateContent() response is empty")
		}
	})

	t.Run("test generate content stream success", func(t *testing.T) {
		var chunks []string
		request := &GenerateContentRequest{
			Messages: []Message{
				{Role: "system", Content: "You are a helpful assistant."},
				{Role: "user", Content: "Count from 1 to 3"},
			},
			Stream: true,
			StreamingFunc: func(_ context.Context, chunk []byte) error {
				chunks = append(chunks, string(chunk))
				return nil
			},
		}

		response, err := client.GenerateContent(ctx, request)
		if err != nil {
			t.Fatalf("GenerateContent() streaming error = %v", err)
		}

		if response == nil {
			t.Fatal("GenerateContent() streaming response is nil")
		}

		// For streaming, we expect chunks to be collected
		if len(chunks) == 0 && rr.Recording() {
			t.Log("Warning: No streaming chunks received during recording")
		}
	})
}
