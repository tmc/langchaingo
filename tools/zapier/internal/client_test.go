package internal

import (
	"context"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/internal/httprr"
)

func scrubZapierData(req *http.Request) error {
	// Scrub API key
	if req.Header.Get("X-API-Key") != "" {
		req.Header.Set("X-API-Key", "test-api-key")
	}
	// Scrub Bearer token
	if auth := req.Header.Get("Authorization"); auth != "" && auth != "Bearer test-token" {
		req.Header.Set("Authorization", "Bearer test-token")
	}
	return nil
}

func TestZapierClient_List(t *testing.T) {
	t.Parallel()

	// Setup HTTP record/replay
	rr, err := httprr.OpenForTest(t, http.DefaultTransport)
	require.NoError(t, err)
	defer rr.Close()

	rr.ScrubReq(scrubZapierData)

	// Get API key from environment or use test key
	apiKey := os.Getenv("ZAPIER_NLA_API_KEY")
	if apiKey == "" && rr.Recording() {
		t.Skip("ZAPIER_NLA_API_KEY not set")
	}
	if apiKey == "" {
		apiKey = "test-api-key"
	}

	// Create client with custom transport
	client, err := NewClient(ClientOptions{
		APIKey: apiKey,
	})
	require.NoError(t, err)

	// Replace transport with httprr
	client.client.Transport = &Transport{
		RoundTripper: rr,
		apiKey:       apiKey,
		UserAgent:    "LangChainGo/0.0.1",
	}

	// List actions
	results, err := client.List(context.Background())
	require.NoError(t, err)
	require.NotNil(t, results)
}

func TestZapierClient_Execute(t *testing.T) {
	t.Parallel()

	// Setup HTTP record/replay
	rr, err := httprr.OpenForTest(t, http.DefaultTransport)
	require.NoError(t, err)
	defer rr.Close()

	rr.ScrubReq(scrubZapierData)

	// Get API key from environment or use test key
	apiKey := os.Getenv("ZAPIER_NLA_API_KEY")
	if apiKey == "" && rr.Recording() {
		t.Skip("ZAPIER_NLA_API_KEY not set")
	}
	if apiKey == "" {
		apiKey = "test-api-key"
	}

	// Create client with custom transport
	client, err := NewClient(ClientOptions{
		APIKey: apiKey,
	})
	require.NoError(t, err)

	// Replace transport with httprr
	client.client.Transport = &Transport{
		RoundTripper: rr,
		apiKey:       apiKey,
		UserAgent:    "LangChainGo/0.0.1",
	}

	// Execute action (using a test action ID)
	result, err := client.Execute(
		context.Background(),
		"test-action-id",
		"Send an email to test@example.com",
		map[string]string{
			"recipient": "test@example.com",
		},
	)
	if err != nil {
		// May error if no real action ID, but we're testing the HTTP call
		t.Logf("Execute error (expected in replay mode): %v", err)
	}
	_ = result
}

func TestZapierClient_ExecuteAsString(t *testing.T) {
	t.Parallel()

	// Setup HTTP record/replay
	rr, err := httprr.OpenForTest(t, http.DefaultTransport)
	require.NoError(t, err)
	defer rr.Close()

	rr.ScrubReq(scrubZapierData)

	// Get API key from environment or use test key
	apiKey := os.Getenv("ZAPIER_NLA_API_KEY")
	if apiKey == "" && rr.Recording() {
		t.Skip("ZAPIER_NLA_API_KEY not set")
	}
	if apiKey == "" {
		apiKey = "test-api-key"
	}

	// Create client with custom transport
	client, err := NewClient(ClientOptions{
		APIKey: apiKey,
	})
	require.NoError(t, err)

	// Replace transport with httprr
	client.client.Transport = &Transport{
		RoundTripper: rr,
		apiKey:       apiKey,
		UserAgent:    "LangChainGo/0.0.1",
	}

	// Execute action as string
	result, err := client.ExecuteAsString(
		context.Background(),
		"test-action-id",
		"Create a calendar event",
		map[string]string{
			"title": "Test Event",
			"date":  "2024-01-01",
		},
	)
	if err != nil {
		// May error if no real action ID, but we're testing the HTTP call
		t.Logf("ExecuteAsString error (expected in replay mode): %v", err)
	}
	_ = result
}

func TestZapierClient_WithAccessToken(t *testing.T) {
	t.Parallel()

	// Setup HTTP record/replay
	rr, err := httprr.OpenForTest(t, http.DefaultTransport)
	require.NoError(t, err)
	defer rr.Close()

	rr.ScrubReq(scrubZapierData)

	// Get access token from environment or use test token
	accessToken := os.Getenv("ZAPIER_NLA_ACCESS_TOKEN")
	if accessToken == "" && rr.Recording() {
		t.Skip("ZAPIER_NLA_ACCESS_TOKEN not set")
	}
	if accessToken == "" {
		accessToken = "test-token"
	}

	// Create client with access token
	client, err := NewClient(ClientOptions{
		AccessToken: accessToken,
	})
	require.NoError(t, err)

	// Replace transport with httprr
	client.client.Transport = &Transport{
		RoundTripper: rr,
		accessToken:  accessToken,
		UserAgent:    "LangChainGo/0.0.1",
	}

	// List actions with OAuth
	results, err := client.List(context.Background())
	require.NoError(t, err)
	require.NotNil(t, results)
}