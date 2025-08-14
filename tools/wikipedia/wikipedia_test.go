package wikipedia

import (
	"context"
	"net/http"
	"testing"

	"github.com/0xDezzy/langchaingo/internal/httprr"
	"github.com/stretchr/testify/require"
)

const _userAgent = "langchaingo test (https://github.com/0xDezzy/langchaingo)"

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
