package ernieclient

import (
	"context"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/internal/httprr"
)

func TestClient_CreateCompletion(t *testing.T) {
	t.Parallel()

	rr, err := httprr.OpenForTest(t, http.DefaultTransport)
	require.NoError(t, err)
	defer rr.Close()

	// Scrub access token from recordings
	rr.ScrubReq(func(req *http.Request) error {
		// Scrub both the access_token in query params and auth headers
		q := req.URL.Query()
		if q.Get("access_token") != "" {
			q.Set("access_token", "test-access-token")
			req.URL.RawQuery = q.Encode()
		}
		return nil
	})

	apiKey := "test-api-key"
	secretKey := "test-secret-key"
	if key := os.Getenv("ERNIE_API_KEY"); key != "" && rr.Recording() {
		apiKey = key
	}
	if secret := os.Getenv("ERNIE_SECRET_KEY"); secret != "" && rr.Recording() {
		secretKey = secret
	}

	client, err := New(
		WithAKSK(apiKey, secretKey),
		WithHTTPClient(rr.Client()),
	)
	require.NoError(t, err)

	req := &CompletionRequest{
		Messages: []Message{
			{
				Role:    "user",
				Content: "你好，请问你是谁？",
			},
		},
		Temperature: 0.7,
	}

	resp, err := client.CreateCompletion(context.Background(), DefaultCompletionModelPath, req)
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp.Result)
}

func TestClient_CreateCompletionStream(t *testing.T) {
	t.Parallel()

	rr, err := httprr.OpenForTest(t, http.DefaultTransport)
	require.NoError(t, err)
	defer rr.Close()

	// Scrub access token from recordings
	rr.ScrubReq(func(req *http.Request) error {
		q := req.URL.Query()
		if q.Get("access_token") != "" {
			q.Set("access_token", "test-access-token")
			req.URL.RawQuery = q.Encode()
		}
		return nil
	})

	apiKey := "test-api-key"
	secretKey := "test-secret-key"
	if key := os.Getenv("ERNIE_API_KEY"); key != "" && rr.Recording() {
		apiKey = key
	}
	if secret := os.Getenv("ERNIE_SECRET_KEY"); secret != "" && rr.Recording() {
		secretKey = secret
	}

	client, err := New(
		WithAKSK(apiKey, secretKey),
		WithHTTPClient(rr.Client()),
	)
	require.NoError(t, err)

	var chunks []string
	req := &CompletionRequest{
		Messages: []Message{
			{
				Role:    "user",
				Content: "数到5",
			},
		},
		Temperature: 0.7,
		Stream:      true,
		StreamingFunc: func(ctx context.Context, chunk []byte) error {
			chunks = append(chunks, string(chunk))
			return nil
		},
	}

	resp, err := client.CreateCompletion(context.Background(), DefaultCompletionModelPath, req)
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, chunks)
}

func TestClient_CreateChat(t *testing.T) {
	t.Parallel()

	rr, err := httprr.OpenForTest(t, http.DefaultTransport)
	require.NoError(t, err)
	defer rr.Close()

	// Scrub access token from recordings
	rr.ScrubReq(func(req *http.Request) error {
		q := req.URL.Query()
		if q.Get("access_token") != "" {
			q.Set("access_token", "test-access-token")
			req.URL.RawQuery = q.Encode()
		}
		return nil
	})

	apiKey := "test-api-key"
	secretKey := "test-secret-key"
	if key := os.Getenv("ERNIE_API_KEY"); key != "" && rr.Recording() {
		apiKey = key
	}
	if secret := os.Getenv("ERNIE_SECRET_KEY"); secret != "" && rr.Recording() {
		secretKey = secret
	}

	client, err := New(
		WithAKSK(apiKey, secretKey),
		WithHTTPClient(rr.Client()),
	)
	require.NoError(t, err)

	req := &ChatRequest{
		Messages: []*ChatMessage{
			{
				Role:    "user",
				Content: "你好",
			},
		},
		Temperature: 0.7,
	}

	resp, err := client.CreateChat(context.Background(), req)
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp.Result)
}

func TestClient_CreateEmbedding(t *testing.T) {
	t.Parallel()

	rr, err := httprr.OpenForTest(t, http.DefaultTransport)
	require.NoError(t, err)
	defer rr.Close()

	// Scrub access token from recordings
	rr.ScrubReq(func(req *http.Request) error {
		q := req.URL.Query()
		if q.Get("access_token") != "" {
			q.Set("access_token", "test-access-token")
			req.URL.RawQuery = q.Encode()
		}
		return nil
	})

	apiKey := "test-api-key"
	secretKey := "test-secret-key"
	if key := os.Getenv("ERNIE_API_KEY"); key != "" && rr.Recording() {
		apiKey = key
	}
	if secret := os.Getenv("ERNIE_SECRET_KEY"); secret != "" && rr.Recording() {
		secretKey = secret
	}

	client, err := New(
		WithAKSK(apiKey, secretKey),
		WithHTTPClient(rr.Client()),
	)
	require.NoError(t, err)

	texts := []string{"你好世界", "今天天气怎么样"}
	resp, err := client.CreateEmbedding(context.Background(), texts)
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp.Data, 2)
	assert.NotEmpty(t, resp.Data[0].Embedding)
	assert.NotEmpty(t, resp.Data[1].Embedding)
}