package maritacaclient

import (
	"context"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vendasta/langchaingo/internal/httprr"
)

func TestClient_Generate(t *testing.T) {
	ctx := context.Background()

	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "MARITACA_API_KEY")

	rr := httprr.OpenForTest(t, http.DefaultTransport)
	defer rr.Close()

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
	err = client.Generate(ctx, req, func(resp ChatResponse) error {
		response = &resp
		return nil
	})
	require.NoError(t, err)
	assert.NotNil(t, response)
	assert.NotEmpty(t, response.Answer)
}

func TestClient_GenerateStream(t *testing.T) {
	ctx := context.Background()

	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "MARITACA_API_KEY")

	rr := httprr.OpenForTest(t, http.DefaultTransport)
	defer rr.Close()

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
	err = client.Generate(ctx, req, func(resp ChatResponse) error {
		responses = append(responses, resp)
		return nil
	})
	require.NoError(t, err)
	assert.NotEmpty(t, responses)
}
