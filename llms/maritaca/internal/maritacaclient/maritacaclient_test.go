package maritacaclient

import (
	"context"
	"net/http"
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

	// Scrub API key from recordings
	rr.ScrubReq(func(req *http.Request) error {
		req.Header.Set("Authorization", "Key test-api-key")
		return nil
	})

	client, err := NewClient(rr.Client())
	require.NoError(t, err)

	apiKey := "test-api-key"
	if key := os.Getenv("MARITACA_API_KEY"); key != "" && rr.Recording() {
		apiKey = key
	}
	client.Token = apiKey
	client.Model = "sabia-2-medium"

	stream := false
	req := &ChatRequest{
		Model: "sabia-2-medium",
		Messages: []*Message{
			{
				Role:    "user",
				Content: "Olá, como você está?",
			},
		},
		Stream: &stream,
		Options: Options{
			Temperature: 0.7,
			MaxTokens:   100,
			DoSample:    true,
		},
	}

	var response *ChatResponse
	err = client.Generate(context.Background(), req, func(resp ChatResponse) error {
		response = &resp
		return nil
	})
	require.NoError(t, err)
	assert.NotNil(t, response)
	assert.NotEmpty(t, response.Answer)
}

func TestClient_GenerateStream(t *testing.T) {
	t.Parallel()

	rr, err := httprr.OpenForTest(t, http.DefaultTransport)
	require.NoError(t, err)
	defer rr.Close()

	// Scrub API key from recordings
	rr.ScrubReq(func(req *http.Request) error {
		req.Header.Set("Authorization", "Key test-api-key")
		return nil
	})

	client, err := NewClient(rr.Client())
	require.NoError(t, err)

	apiKey := "test-api-key"
	if key := os.Getenv("MARITACA_API_KEY"); key != "" && rr.Recording() {
		apiKey = key
	}
	client.Token = apiKey
	client.Model = "sabia-2-medium"

	stream := true
	req := &ChatRequest{
		Model: "sabia-2-medium",
		Messages: []*Message{
			{
				Role:    "user",
				Content: "Conte de 1 a 5",
			},
		},
		Stream: &stream,
		Options: Options{
			Temperature: 0.7,
			MaxTokens:   50,
			DoSample:    true,
		},
	}

	var responses []ChatResponse
	err = client.Generate(context.Background(), req, func(resp ChatResponse) error {
		responses = append(responses, resp)
		return nil
	})
	require.NoError(t, err)
	assert.NotEmpty(t, responses)
}