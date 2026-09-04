package googleai

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type redirectTransport struct{ host string }

func (t redirectTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	to, err := url.Parse(t.host)
	if err != nil {
		return nil, err
	}
	req = req.Clone(req.Context())
	req.URL.Scheme, req.URL.Host, req.Host = to.Scheme, to.Host, to.Host
	return http.DefaultTransport.RoundTrip(req)
}

func TestListModelsStripsTheResourcePrefix(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"models": []map[string]any{
				{"name": "models/gemini-3.8-flash"},
				{"name": "models/gemma-4-31b-it"},
				{"name": ""},
			},
		})
	}))
	defer srv.Close()

	llm, err := New(t.Context(), WithRest(), WithAPIKey("test-key"),
		WithHTTPClient(&http.Client{Transport: redirectTransport{host: srv.URL}}))
	require.NoError(t, err)

	ids, err := llm.ListModels(t.Context())
	require.NoError(t, err)
	assert.Equal(t, []string{"gemini-3.8-flash", "gemma-4-31b-it"}, ids)
}
