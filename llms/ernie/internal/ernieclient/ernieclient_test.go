package ernieclient

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

func requireErnieCredentialsOrHTTPRR(t *testing.T) *httprr.RecordReplay {
	t.Helper()

	// Check if we have API credentials or httprr recording
	hasCredentials := os.Getenv("ERNIE_API_KEY") != "" && os.Getenv("ERNIE_SECRET_KEY") != ""

	if !hasCredentials {
		testName := httprr.CleanFileName(t.Name())
		httprrFile := filepath.Join("testdata", testName+".httprr")
		httprrGzFile := httprrFile + ".gz"
		if _, err := os.Stat(httprrFile); os.IsNotExist(err) {
			if _, err := os.Stat(httprrGzFile); os.IsNotExist(err) {
				t.Skip("ERNIE_API_KEY and ERNIE_SECRET_KEY not set and no httprr recording available")
			}
		}
	}

	rr := httprr.OpenForTest(t, http.DefaultTransport)
	return rr
}

func TestClient_CreateCompletion(t *testing.T) {
	ctx := context.Background()
	t.Parallel()

	rr := requireErnieCredentialsOrHTTPRR(t)
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

	apiKey := os.Getenv("ERNIE_API_KEY")
	if apiKey == "" {
		apiKey = "test-api-key"
	}
	secretKey := os.Getenv("ERNIE_SECRET_KEY")
	if secretKey == "" {
		secretKey = "test-secret-key"
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

	resp, err := client.CreateCompletion(ctx, DefaultCompletionModelPath, req)
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp.Result)
}

func TestClient_CreateCompletionStream(t *testing.T) {
	ctx := context.Background()
	t.Parallel()

	rr := requireErnieCredentialsOrHTTPRR(t)
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
	apiKey := os.Getenv("ERNIE_API_KEY")
	if apiKey == "" {
		apiKey = "test-api-key"
	}
	secretKey := os.Getenv("ERNIE_SECRET_KEY")
	if secretKey == "" {
		secretKey = "test-secret-key"
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

	resp, err := client.CreateCompletion(ctx, DefaultCompletionModelPath, req)
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, chunks)
}

func TestClient_CreateChat(t *testing.T) {
	ctx := context.Background()
	t.Parallel()

	rr := requireErnieCredentialsOrHTTPRR(t)
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
	apiKey := os.Getenv("ERNIE_API_KEY")
	if apiKey == "" {
		apiKey = "test-api-key"
	}
	secretKey := os.Getenv("ERNIE_SECRET_KEY")
	if secretKey == "" {
		secretKey = "test-secret-key"
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

	resp, err := client.CreateChat(ctx, req)
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp.Result)
}

func TestClient_CreateEmbedding(t *testing.T) {
	ctx := context.Background()
	t.Parallel()

	rr := requireErnieCredentialsOrHTTPRR(t)
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
	apiKey := os.Getenv("ERNIE_API_KEY")
	if apiKey == "" {
		apiKey = "test-api-key"
	}
	secretKey := os.Getenv("ERNIE_SECRET_KEY")
	if secretKey == "" {
		secretKey = "test-secret-key"
	}

	client, err := New(
		WithAKSK(apiKey, secretKey),
		WithHTTPClient(rr.Client()),
	)
	require.NoError(t, err)

	texts := []string{"你好世界", "今天天气怎么样"}
	resp, err := client.CreateEmbedding(ctx, texts)
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp.Data, 2)
	assert.NotEmpty(t, resp.Data[0].Embedding)
	assert.NotEmpty(t, resp.Data[1].Embedding)
}
