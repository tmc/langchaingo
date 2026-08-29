package mistral

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mistralAnswering(t *testing.T, status int, body string) *Model {
	t.Helper()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, r.Body)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = io.WriteString(w, body)
	}))
	t.Cleanup(srv.Close)

	llm, err := New(WithAPIKey("test"), WithEndpoint(srv.URL), WithModel("mistral-small-latest"))
	require.NoError(t, err)
	return llm
}

func TestCallReportsAFailedRequestToTheHandler(t *testing.T) {
	t.Parallel()

	llm := mistralAnswering(t, http.StatusInternalServerError, `{"error":"upstream is down"}`)
	rec := &errorRecorder{}
	llm.CallbacksHandler = rec

	out, err := llm.Call(context.Background(), "hi")

	require.Error(t, err)
	assert.Empty(t, out)
	require.Len(t, rec.errs, 1, "a failed request must reach the handler exactly once")
	assert.ErrorIs(t, rec.errs[0], err)
}

func TestCallRejectsAnAnswerThatIsNotASingleChoice(t *testing.T) {
	t.Parallel()

	llm := mistralAnswering(t, http.StatusOK,
		`{"id":"x","object":"chat.completion","model":"mistral-small-latest","choices":[],`+
			`"usage":{"prompt_tokens":1,"completion_tokens":0,"total_tokens":1}}`)
	rec := &errorRecorder{}
	llm.CallbacksHandler = rec

	out, err := llm.Call(context.Background(), "hi")

	require.Error(t, err)
	assert.Empty(t, out)
	assert.Contains(t, err.Error(), "length of the Choices slice must be 1")
	require.Len(t, rec.errs, 1)
	assert.ErrorIs(t, rec.errs[0], err)
}
