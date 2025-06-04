package serpapi

import (
	"context"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/tmc/langchaingo/internal/httprr"
)

func TestSerpAPITool(t *testing.T) {
	t.Parallel()

	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "SERPAPI_API_KEY")

	rr := httprr.OpenForTest(t, http.DefaultTransport)
	t.Cleanup(func() { rr.Close() })

	// Scrub the API key from requests
	rr.ScrubReq(func(req *http.Request) error {
		if req.URL != nil {
			q := req.URL.Query()
			q.Set("api_key", "test-api-key")
			req.URL.RawQuery = q.Encode()
		}
		return nil
	})

	// Create tool - use dummy key for replay mode, real key for recording
	var apiKey string
	if rr.Recording() {
		apiKey = os.Getenv("SERPAPI_API_KEY")
	} else {
		apiKey = "test-api-key"
	}

	// Create tool with HTTP client
	tool, err := New(
		WithAPIKey(apiKey),
		WithHTTPClient(rr.Client()),
	)
	if err != nil {
		t.Fatalf("Failed to create tool: %v", err)
	}

	// Test search functionality with a stable question
	result, err := tool.Call(context.Background(), "What year was Unix first released at Bell Labs?")
	if err != nil {
		t.Fatalf("Tool call failed: %v", err)
	}
	if result == "" {
		t.Fatal("Result should not be empty")
	}

	// Basic validation - result should contain meaningful content
	if len(result) <= 3 {
		t.Errorf("Result should contain more than 3 characters, got: %s", result)
	}

	// For debugging test failures, log the result
	t.Logf("SerpAPI result: %s", result)
}

func TestSerpAPIToolError(t *testing.T) {
	t.Parallel()

	// Save original environment variable
	originalKey := os.Getenv("SERPAPI_API_KEY")
	// Temporarily unset the environment variable
	os.Setenv("SERPAPI_API_KEY", "")
	// Restore it after the test
	t.Cleanup(func() {
		os.Setenv("SERPAPI_API_KEY", originalKey)
	})

	// Test error handling when no API key is provided
	_, err := New()
	if err == nil {
		t.Fatal("Expected error when no API key is provided")
	}
	if !strings.Contains(err.Error(), "missing the serpapi API key") {
		t.Errorf("Expected missing API key error, got: %v", err)
	}
}
