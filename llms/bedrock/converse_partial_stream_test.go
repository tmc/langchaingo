package bedrock_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/llms/bedrock"
	"github.com/vxcontrol/langchaingo/llms/streaming"
)

func TestConverseStreamKeepsWhatArrivedWhenTheConsumerGivesUp(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, r.Body)
		w.Header().Set("Content-Type", "application/vnd.amazon.eventstream")
		enc := eventstream.NewEncoder()
		writeConverseEvent(t, w, enc, "messageStart", `{"role":"assistant"}`)
		writeConverseEvent(t, w, enc, "contentBlockDelta",
			`{"contentBlockIndex":0,"delta":{"text":"sixty "}}`)
		writeConverseEvent(t, w, enc, "contentBlockDelta",
			`{"contentBlockIndex":0,"delta":{"text":"rooms "}}`)
		writeConverseEvent(t, w, enc, "contentBlockDelta",
			`{"contentBlockIndex":0,"delta":{"text":"are free"}}`)
		writeConverseEvent(t, w, enc, "contentBlockStop", `{"contentBlockIndex":0}`)
		writeConverseEvent(t, w, enc, "messageStop", `{"stopReason":"end_turn"}`)
		writeConverseEvent(t, w, enc, "metadata",
			`{"usage":{"inputTokens":10,"outputTokens":9,"totalTokens":19}}`)
	}))
	t.Cleanup(srv.Close)

	llm := bedrockLLMAgainst(t, srv,
		bedrock.WithModel("anthropic.claude-sonnet-4-5-20250929-v1:0"), bedrock.WithConverseAPI())

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
	assert.Equal(t, "sixty rooms ", resp.Choices[0].Content)
}
