package openai

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

func TestAConsumerThatGivesUpStillGetsWhatArrived(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, r.Body)
		w.Header().Set("Content-Type", "text/event-stream")
		for _, text := range []string{"sixty ", "rooms ", "are free"} {
			_, _ = io.WriteString(w, `data: {"id":"x","object":"chat.completion.chunk","created":1,`+
				`"model":"gpt-4o","choices":[{"index":0,"delta":{"content":"`+text+`"},"finish_reason":null}]}`+"\n\n")
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
		}
		_, _ = io.WriteString(w, "data: [DONE]\n\n")
	}))
	t.Cleanup(srv.Close)

	llm := newUnitLLM(t, WithBaseURL(srv.URL), WithModel("gpt-4o"))

	gaveUp := errors.New("consumer gave up")
	delivered := 0
	resp, err := llm.GenerateContent(context.Background(),
		[]llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "how many rooms are free?")},
		llms.WithStreamingFunc(func(_ context.Context, chunk streaming.Chunk) error {
			if chunk.Type != streaming.ChunkTypeText {
				return nil
			}
			delivered++
			if delivered == 2 {
				return gaveUp
			}
			return nil
		}))

	require.ErrorIs(t, err, gaveUp)
	require.NotNil(t, resp, "the text already delivered must travel with the error")
	require.Len(t, resp.Choices, 1)
	assert.Equal(t, "sixty rooms ", resp.Choices[0].Content,
		"every chunk written into the accumulator before the callback failed")
}
