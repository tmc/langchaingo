package mistral

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func embeddingModelOnTheWire(t *testing.T, opts ...Option) string {
	t.Helper()

	var got string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		var body struct {
			Model string `json:"model"`
		}
		require.NoError(t, json.Unmarshal(b, &body))
		got = body.Model
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[{"embedding":[0.1,0.2]}]}`))
	}))
	t.Cleanup(srv.Close)

	m, err := New(append([]Option{WithEndpoint(srv.URL), WithAPIKey("k")}, opts...)...)
	require.NoError(t, err)
	_, err = m.CreateEmbedding(context.Background(), []string{"x"})
	require.NoError(t, err)
	return got
}

func TestTheCallersEmbeddingModelReachesTheWire(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "mistral-embed", embeddingModelOnTheWire(t),
		"a caller that named no model still embeds, and not with the chat default")
	assert.Equal(t, "mistral/mistral-embed", embeddingModelOnTheWire(t, WithEmbeddingModel("mistral/mistral-embed")),
		"the named model must reach the wire: a gateway routes by the prefixed name")
	assert.Equal(t, "mistral-embed", embeddingModelOnTheWire(t, WithModel("ministral-8b-latest")),
		"WithModel names the chat model and must not follow the caller into embeddings")
}
