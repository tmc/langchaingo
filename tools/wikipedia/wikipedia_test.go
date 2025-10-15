package wikipedia

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vendasta/langchaingo/internal/httprr"
)

const _userAgent = "langchaingo test (https://github.com/vendasta/langchaingo)"

func TestWikipedia(t *testing.T) {
	ctx := context.Background()
	t.Parallel()

	// Setup httprr for HTTP requests
	rr := httprr.OpenForTest(t, http.DefaultTransport)
	t.Cleanup(func() { rr.Close() })

	tool := New(_userAgent, WithHTTPClient(rr.Client()))
	_, err := tool.Call(ctx, "america")
	require.NoError(t, err)
}
