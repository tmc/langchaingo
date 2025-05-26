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
	t.Parallel()

	// Setup HTTP record/replay for Wikipedia API calls
	rr, err := httprr.OpenForTest(t, http.DefaultTransport)
	require.NoError(t, err)
	defer rr.Close()

	tool := NewWithClient(_userAgent, rr.Client())
	_, err = tool.Call(context.Background(), "america")
	require.NoError(t, err)
}
