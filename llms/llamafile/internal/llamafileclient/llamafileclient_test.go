package llamafileclient

import (
	"context"
	"net/http"
	"net/url"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/internal/httprr"
)

func TestClient_Generate(t *testing.T) {
	t.Parallel()

	rr, err := httprr.OpenForTest(t, http.DefaultTransport)
	require.NoError(t, err)
	defer rr.Close()

	baseURL := "http://localhost:8080"
	if envURL := os.Getenv("LLAMAFILE_HOST"); envURL != "" && rr.Recording() {
		baseURL = envURL
	}

	parsedURL, err := url.Parse(baseURL)
	require.NoError(t, err)

	client, err := NewClient(parsedURL, rr.Client())
	require.NoError(t, err)

	stream := false
	req := &GenerateRequest{
		Prompt: "Hello, how are you?",
		Stream: &stream,
		GenerationSettings: GenerationSettings{
			Temperature: 0.7,
			NPredict:    100,
		},
	}

	var response *GenerateResponse
	err = client.Generate(context.Background(), req, func(resp GenerateResponse) error {
		response = &resp
		return nil
	})
	require.NoError(t, err)
	assert.NotNil(t, response)
	assert.NotEmpty(t, response.Response)
	assert.True(t, response.Done)
}

func TestClient_GenerateStream(t *testing.T) {
	t.Parallel()

	rr, err := httprr.OpenForTest(t, http.DefaultTransport)
	require.NoError(t, err)
	defer rr.Close()

	baseURL := "http://localhost:8080"
	if envURL := os.Getenv("LLAMAFILE_HOST"); envURL != "" && rr.Recording() {
		baseURL = envURL
	}

	parsedURL, err := url.Parse(baseURL)
	require.NoError(t, err)

	client, err := NewClient(parsedURL, rr.Client())
	require.NoError(t, err)

	stream := true
	req := &GenerateRequest{
		Prompt: "Count from 1 to 5",
		Stream: &stream,
		GenerationSettings: GenerationSettings{
			Temperature: 0.7,
			NPredict:    50,
		},
	}

	var responses []GenerateResponse
	err = client.Generate(context.Background(), req, func(resp GenerateResponse) error {
		responses = append(responses, resp)
		return nil
	})
	require.NoError(t, err)
	assert.NotEmpty(t, responses)
	assert.True(t, responses[len(responses)-1].Done)
}

func TestClient_GenerateChat(t *testing.T) {
	t.Parallel()

	rr, err := httprr.OpenForTest(t, http.DefaultTransport)
	require.NoError(t, err)
	defer rr.Close()

	baseURL := "http://localhost:8080"
	if envURL := os.Getenv("LLAMAFILE_HOST"); envURL != "" && rr.Recording() {
		baseURL = envURL
	}

	parsedURL, err := url.Parse(baseURL)
	require.NoError(t, err)

	client, err := NewClient(parsedURL, rr.Client())
	require.NoError(t, err)

	stream := false
	req := &ChatRequest{
		Messages: []*Message{
			{
				Role:    "user",
				Content: "Hello, how are you?",
			},
		},
		Stream: &stream,
		GenerationSettings: GenerationSettings{
			Temperature: 0.7,
			NPredict:    50,
		},
	}

	var response *ChatResponse
	err = client.GenerateChat(context.Background(), req, func(resp ChatResponse) error {
		response = &resp
		return nil
	})
	require.NoError(t, err)
	assert.NotNil(t, response)
	assert.NotEmpty(t, response.Content)
}

func TestClient_CreateEmbedding(t *testing.T) {
	t.Parallel()

	rr, err := httprr.OpenForTest(t, http.DefaultTransport)
	require.NoError(t, err)
	defer rr.Close()

	baseURL := "http://localhost:8080"
	if envURL := os.Getenv("LLAMAFILE_HOST"); envURL != "" && rr.Recording() {
		baseURL = envURL
	}

	parsedURL, err := url.Parse(baseURL)
	require.NoError(t, err)

	client, err := NewClient(parsedURL, rr.Client())
	require.NoError(t, err)

	texts := []string{"Hello world", "How are you?"}
	resp, err := client.CreateEmbedding(context.Background(), texts)
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp.Results, 2)
	assert.NotEmpty(t, resp.Results[0].Embedding)
	assert.NotEmpty(t, resp.Results[1].Embedding)
}