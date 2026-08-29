package ollama

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/llms/streaming"
)

func thinkingOnlyStream(t *testing.T) *LLM {
	t.Helper()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, r.Body)
		w.Header().Set("Content-Type", "application/x-ndjson")
		for _, thought := range []string{"counting the ", "free rooms"} {
			_, _ = io.WriteString(w, `{"model":"llama3","created_at":"2026-08-21T09:00:00Z",`+
				`"message":{"role":"assistant","content":"","thinking":"`+thought+`"},"done":false}`+"\n")
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
		}
	}))
	t.Cleanup(srv.Close)

	llm, err := New(WithServerURL(srv.URL), WithModel("llama3"))
	require.NoError(t, err)
	return llm
}

func TestAReasoningOnlyStreamIsNotThrownAway(t *testing.T) {
	t.Parallel()

	llm := thinkingOnlyStream(t)

	gaveUp := errors.New("consumer gave up")
	delivered := 0
	resp, err := llm.GenerateContent(context.Background(),
		[]llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "how many rooms are free?")},
		llms.WithStreamingFunc(func(_ context.Context, chunk streaming.Chunk) error {
			if chunk.Type != streaming.ChunkTypeReasoning {
				return nil
			}
			delivered++
			if delivered == 2 {
				return gaveUp
			}
			return nil
		}))

	require.ErrorIs(t, err, gaveUp)
	require.NotNil(t, resp, "a stream that produced only reasoning still produced something")
	require.Len(t, resp.Choices, 1)
	require.NotNil(t, resp.Choices[0].Reasoning)
	assert.Contains(t, resp.Choices[0].Reasoning.Content, "counting the ")
}
