package duckduckgo

import (
	"context"
	"net/http"
	"testing"

	"github.com/0xDezzy/langchaingo/internal/httprr"
	"github.com/stretchr/testify/require"
)

func TestDuckDuckGoTool(t *testing.T) {
	ctx := context.Background()
	t.Parallel()

	// Setup httprr for HTTP requests
	rr := httprr.OpenForTest(t, http.DefaultTransport)
	t.Cleanup(func() { rr.Close() })

	// Create tool with httprr HTTP client
	tool, err := New(3, DefaultUserAgent, WithHTTPClient(rr.Client()))
	require.NoError(t, err)

	// Test search functionality
	result, err := tool.Call(ctx, "golang programming language")
	require.NoError(t, err)
	require.NotEmpty(t, result)

	// Basic validation - should contain some search results
	require.True(t, len(result) > 10, "Result should contain meaningful content")
	require.Contains(t, result, "Title:", "Result should contain formatted search results")
}

func TestDuckDuckGoToolBasicConstruction(t *testing.T) {
	t.Parallel()

	// Test basic construction
	tool, err := New(5, DefaultUserAgent)
	require.NoError(t, err)
	require.NotNil(t, tool)
	require.Equal(t, "DuckDuckGo Search", tool.Name())
	require.NotEmpty(t, tool.Description())
}
