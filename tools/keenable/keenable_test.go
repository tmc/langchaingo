package keenable

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/httputil"
	"github.com/tmc/langchaingo/internal/httprr"
)

func TestKeenableTool(t *testing.T) {
	ctx := context.Background()

	// The keyless public endpoint needs no credentials, so the recording can
	// always be refreshed with -httprecord.
	rr := httprr.OpenForTest(t, httputil.DefaultTransport)
	if rr.Replaying() {
		t.Parallel()
	}

	tool, err := New(WithAPIKey(""), WithMaxResults(3), WithHTTPClient(rr.Client()))
	require.NoError(t, err)

	result, err := tool.Call(ctx, "golang programming language")
	require.NoError(t, err)
	require.Contains(t, result, "Title:")
	require.Contains(t, result, "URL: https://")
}

func TestKeenableToolKeylessRequest(t *testing.T) {
	t.Parallel()

	var (
		gotPath, gotTitle, gotKey string
		gotBody                   map[string]any
	)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotTitle = r.Header.Get("X-Keenable-Title")
		gotKey = r.Header.Get("X-API-Key")
		require.NoError(t, json.NewDecoder(r.Body).Decode(&gotBody))
		_, _ = w.Write([]byte(`{"results":[` +
			`{"title":"Go","url":"https://go.dev","description":"","snippet":"Build simple, secure, scalable systems with Go.","published_at":"2026-05-06T10:00:00Z"},` +
			`{"title":"no url","url":"","snippet":"dropped"}]}`))
	}))
	t.Cleanup(server.Close)

	tool, err := New(WithAPIKey(""), WithBaseURL(server.URL+"/v1/search"), WithMaxResults(2), WithSite("go.dev"))
	require.NoError(t, err)
	result, err := tool.Call(context.Background(), "go")
	require.NoError(t, err)

	require.Equal(t, "/v1/search/public", gotPath)
	require.Equal(t, DefaultAppTitle, gotTitle)
	require.Empty(t, gotKey)
	require.Equal(t, map[string]any{"query": "go", "max_results": float64(2), "site": "go.dev"}, gotBody)
	require.Equal(t, "Title: Go\nPublished: 2026-05-06T10:00:00Z\nDescription: Build simple, secure, scalable systems with Go.\nURL: https://go.dev\n\n", result)
}

func TestKeenableToolKeyedRequest(t *testing.T) {
	t.Parallel()

	var gotPath, gotTitle, gotKey string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotTitle = r.Header.Get("X-Keenable-Title")
		gotKey = r.Header.Get("X-API-Key")
		_, _ = w.Write([]byte(`{"results":[]}`))
	}))
	t.Cleanup(server.Close)

	tool, err := New(WithAPIKey("test-api-key"), WithBaseURL(server.URL+"/v1/search/"))
	require.NoError(t, err)
	result, err := tool.Call(context.Background(), "go")
	require.NoError(t, err)

	require.Equal(t, "/v1/search", gotPath)
	require.Equal(t, "test-api-key", gotKey)
	require.Empty(t, gotTitle)
	require.Equal(t, "No good Keenable Search Results was found", result)
}

func TestKeenableToolHTTPError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"detail":"rate limited"}`))
	}))
	t.Cleanup(server.Close)

	tool, err := New(WithAPIKey(""), WithBaseURL(server.URL+"/v1/search"))
	require.NoError(t, err)
	_, err = tool.Call(context.Background(), "go")
	require.ErrorContains(t, err, "429")
}

func TestKeenableToolConstruction(t *testing.T) {
	t.Parallel()

	tool, err := New(WithMaxResults(500))
	require.NoError(t, err)
	require.Equal(t, "Keenable Search", tool.Name())
	require.NotEmpty(t, tool.Description())
	require.Equal(t, maxResultsLimit, tool.maxResults)

	tool, err = New(WithMaxResults(0))
	require.NoError(t, err)
	require.Equal(t, 1, tool.maxResults)
}
