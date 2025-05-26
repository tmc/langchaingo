package duckduckgo

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/internal/httprr"
)

func TestDuckDuckGoTool(t *testing.T) {
	t.Parallel()

	// Setup HTTP record/replay for DuckDuckGo calls
	rr, err := httprr.OpenForTest(t, http.DefaultTransport)
	require.NoError(t, err)
	defer rr.Close()

	// Create tool with httprr client
	tool, err := NewWithClient(3, DefaultUserAgent, rr.Client())
	require.NoError(t, err)

	// Test search functionality
	result, err := tool.Call(context.Background(), "golang programming language")
	require.NoError(t, err)
	require.NotEmpty(t, result)
	
	// Basic validation - should contain some search results
	require.True(t, len(result) > 10, "Result should contain meaningful content")
	require.Contains(t, result, "Title:", "Result should contain formatted search results")
}

func TestDuckDuckGoToolBasicConstruction(t *testing.T) {
	t.Parallel()

	// Test basic construction without HTTP client
	tool, err := New(5, DefaultUserAgent)
	require.NoError(t, err)
	require.NotNil(t, tool)
	require.Equal(t, "DuckDuckGo Search", tool.Name())
	require.NotEmpty(t, tool.Description())
}