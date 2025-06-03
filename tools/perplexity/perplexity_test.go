package perplexity

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/internal/httprr"
)

func TestTool_Integration(t *testing.T) {
	ctx := context.Background()
	t.Parallel()

	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "PERPLEXITY_API_KEY")

	rr := httprr.OpenForTest(t, http.DefaultTransport)
	t.Cleanup(func() { rr.Close() })

	tool, err := New(WithHTTPClient(rr.Client()))
	require.NoError(t, err)
	require.NotNil(t, tool)

	assert.Equal(t, "PerplexityAI", tool.Name())
	assert.NotEmpty(t, tool.Description())

	// Test Call functionality
	response, err := tool.Call(ctx, "what is the largest country in the world by total area?")
	require.NoError(t, err)
	assert.Contains(t, response, "Russia")
}
