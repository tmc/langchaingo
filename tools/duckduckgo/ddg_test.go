package duckduckgo

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDuckDuckGoTool(t *testing.T) {
	ctx := context.Background()
	t.Parallel()

	// Create tool
	tool, err := New(3, DefaultUserAgent)
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

	// Test basic construction without HTTP client
	tool, err := New(5, DefaultUserAgent)
	require.NoError(t, err)
	require.NotNil(t, tool)
	require.Equal(t, "DuckDuckGo Search", tool.Name())
	require.NotEmpty(t, tool.Description())
}
