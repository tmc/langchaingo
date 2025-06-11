package perplexity

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/internal/httprr"
)

func TestPerplexityTool(t *testing.T) {
	ctx := context.Background()

	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "PERPLEXITY_API_KEY")

	rr := httprr.OpenForTest(t, http.DefaultTransport)

	var opts []Option
	opts = append(opts, WithHTTPClient(rr.Client()))

	// Use test token when replaying
	if rr.Replaying() {
		opts = append(opts, WithAPIKey("test-api-key"))
	}

	tool, err := New(opts...)
	require.NoError(t, err)
	require.NotNil(t, tool)

	assert.Equal(t, "PerplexityAI", tool.Name())
	assert.NotEmpty(t, tool.Description())

	// Test Call functionality
	response, err := tool.Call(ctx, "what is the largest country in the world by total area?")
	require.NoError(t, err)
	assert.Contains(t, response, "Russia")
}
