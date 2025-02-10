package perplexity

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTool_Integration(t *testing.T) {
	t.Parallel()

	apiKey := os.Getenv("PERPLEXITY_API_KEY")
	if apiKey == "" {
		t.Skip("PERPLEXITY_API_KEY not set")
	}

	tool, err := New()
	require.NoError(t, err)
	require.NotNil(t, tool)

	assert.Equal(t, "PerplexityAI", tool.Name())
	assert.NotEmpty(t, tool.Description())

	// Test Call functionality
	ctx := context.Background()
	response, err := tool.Call(ctx, "what is the largest country in the world by total area?")
	require.NoError(t, err)
	assert.Contains(t, response, "Russia")
}
