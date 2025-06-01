package wikipedia

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/internal/httprr"
)

const _userAgent = "langchaingo test (https://github.com/tmc/langchaingo)"

func TestWikipedia(t *testing.T) {
	ctx := context.Background()

	// Setup httprr for HTTP requests
	rr := httprr.OpenForTest(t, http.DefaultTransport)

	tool := New(_userAgent, WithHTTPClient(rr.Client()))
	_, err := tool.Call(ctx, "america")
	require.NoError(t, err)
}
