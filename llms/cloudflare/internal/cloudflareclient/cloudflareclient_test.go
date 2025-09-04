package cloudflareclient

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/internal/httprr"
)

const testBaseURL = "https://api.cloudflare.com/client/v4/accounts"

func requireCloudflareCredentialsOrHTTPRR(t *testing.T) *httprr.RecordReplay {
	t.Helper()

	// Check if we have API credentials or httprr recording
	hasCredentials := os.Getenv("CLOUDFLARE_ACCOUNT_ID") != "" && os.Getenv("CLOUDFLARE_API_KEY") != ""

	if !hasCredentials {
		testName := httprr.CleanFileName(t.Name())
		httprrFile := filepath.Join("testdata", testName+".httprr")
		httprrGzFile := httprrFile + ".gz"
		if _, err := os.Stat(httprrFile); os.IsNotExist(err) {
			if _, err := os.Stat(httprrGzFile); os.IsNotExist(err) {
				t.Skip("CLOUDFLARE_ACCOUNT_ID and CLOUDFLARE_API_KEY not set and no httprr recording available")
			}
		}
	}

	rr := httprr.OpenForTest(t, http.DefaultTransport)
	return rr
}

func TestClient_GenerateContentWithHTTPRR(t *testing.T) {
	ctx := context.Background()
	t.Parallel()

	rr := requireCloudflareCredentialsOrHTTPRR(t)
	defer rr.Close()

	accountID := "test-account-id"
	apiKey := "test-api-key"
	model := "@cf/meta/llama-2-7b-chat-int8"

	if id := os.Getenv("CLOUDFLARE_ACCOUNT_ID"); id != "" && rr.Recording() {
		accountID = id
	}
	if key := os.Getenv("CLOUDFLARE_API_KEY"); key != "" && rr.Recording() {
		apiKey = key
	}

	client := NewClient(rr.Client(), accountID, testBaseURL, apiKey, model, "")

	req := &GenerateContentRequest{
		Messages: []Message{
			{
				Role:    RoleTypeUser,
				Content: "Hello, how are you?",
			},
		},
		Stream: false,
	}

	resp, err := client.GenerateContent(ctx, req)
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.True(t, resp.Success)
	assert.NotEmpty(t, resp.Result.Response)
}

func TestClient_GenerateContentStream(t *testing.T) {
	ctx := context.Background()
	t.Parallel()

	rr := requireCloudflareCredentialsOrHTTPRR(t)
	defer rr.Close()

	accountID := "test-account-id"
	apiKey := "test-api-key"
	model := "@cf/meta/llama-2-7b-chat-int8"

	if id := os.Getenv("CLOUDFLARE_ACCOUNT_ID"); id != "" && rr.Recording() {
		accountID = id
	}
	if key := os.Getenv("CLOUDFLARE_API_KEY"); key != "" && rr.Recording() {
		apiKey = key
	}

	client := NewClient(rr.Client(), accountID, testBaseURL, apiKey, model, "")

	var chunks []string
	req := &GenerateContentRequest{
		Messages: []Message{
			{
				Role:    RoleTypeUser,
				Content: "Count from 1 to 5",
			},
		},
		Stream: true,
		StreamingFunc: func(ctx context.Context, chunk []byte) error {
			chunks = append(chunks, string(chunk))
			return nil
		},
	}

	resp, err := client.GenerateContent(ctx, req)
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, chunks)
}

func TestClient_CreateEmbedding(t *testing.T) {
	ctx := context.Background()
	t.Parallel()

	rr := requireCloudflareCredentialsOrHTTPRR(t)
	defer rr.Close()

	accountID := "test-account-id"
	apiKey := "test-api-key"
	embeddingModel := "@cf/baai/bge-base-en-v1.5"

	if id := os.Getenv("CLOUDFLARE_ACCOUNT_ID"); id != "" && rr.Recording() {
		accountID = id
	}
	if key := os.Getenv("CLOUDFLARE_API_KEY"); key != "" && rr.Recording() {
		apiKey = key
	}

	// For embeddings, we need to set the model as the embedding model in the client
	client := NewClient(rr.Client(), accountID, testBaseURL, apiKey, embeddingModel, embeddingModel)

	req := &CreateEmbeddingRequest{
		Text: []string{"Hello world", "How are you?"},
	}

	resp, err := client.CreateEmbedding(ctx, req)
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp.Result.Data)
	assert.Len(t, resp.Result.Data, 2)
	assert.NotEmpty(t, resp.Result.Data[0])
	assert.NotEmpty(t, resp.Result.Data[1])
}
