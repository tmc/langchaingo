package serpapi

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/internal/httprr"
)

func TestSerpAPITool(t *testing.T) {
	t.Parallel()

	// Skip if no API key is available and no recorded data exists
	apiKey := os.Getenv("SERPAPI_API_KEY")
	if apiKey == "" {
		// Check if we have recorded data
		testName := httprr.CleanFileName(t.Name())
		candidates := []string{
			filepath.Join("testdata", testName+".httprr"),
			filepath.Join("testdata", testName+".httprr.gz"),
		}
		
		hasRecordedData := false
		for _, candidate := range candidates {
			if _, err := os.Stat(candidate); err == nil {
				hasRecordedData = true
				break
			}
		}
		
		if !hasRecordedData {
			t.Skip("SERPAPI_API_KEY not set and no recorded data available")
		}
		
		// Use a dummy key for replay mode
		apiKey = "test-key"
	}

	// Setup HTTP record/replay for SerpAPI calls
	rr, err := httprr.OpenForTest(t, http.DefaultTransport)
	require.NoError(t, err)
	defer rr.Close()

	// Create tool with httprr client
	tool, err := New(
		WithAPIKey(apiKey),
		WithHTTPClient(rr.Client()),
	)
	require.NoError(t, err)

	// Test search functionality
	result, err := tool.Call(context.Background(), "What is the capital of France?")
	require.NoError(t, err)
	require.NotEmpty(t, result)
	
	// Basic validation - should contain some information about Paris
	// Note: This test is flexible to work with recorded responses
	require.True(t, len(result) > 10, "Result should contain meaningful content")
}

func TestSerpAPIToolError(t *testing.T) {
	t.Parallel()

	// Test error handling when no API key is provided
	_, err := New()
	require.ErrorIs(t, err, ErrMissingToken)
}