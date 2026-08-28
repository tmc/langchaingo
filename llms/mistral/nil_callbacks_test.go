package mistral

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCallWithoutACallbacksHandlerDoesNotPanic(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, r.Body)
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = io.WriteString(w, `{"error":"upstream is down"}`)
	}))
	t.Cleanup(srv.Close)

	llm, err := New(WithAPIKey("test"), WithEndpoint(srv.URL), WithModel("mistral-small-latest"))
	require.NoError(t, err)
	llm.CallbacksHandler = nil

	require.NotPanics(t, func() {
		_, _ = llm.Call(context.Background(), "hi")
	})
}
