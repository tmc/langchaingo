package bedrock_test

import (
	"context"
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

func writeConverseEvent(t *testing.T, w io.Writer, enc *eventstream.Encoder, event, payload string) {
	t.Helper()

	require.NoError(t, enc.Encode(w, eventstream.Message{
		Headers: eventstream.Headers{
			{Name: ":message-type", Value: eventstream.StringValue("event")},
			{Name: ":event-type", Value: eventstream.StringValue(event)},
			{Name: ":content-type", Value: eventstream.StringValue("application/json")},
		},
		Payload: []byte(payload),
	}))
}

func converseStreamLLM(t *testing.T, stopReason string) *bedrock.LLM {
	t.Helper()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, r.Body)
		w.Header().Set("Content-Type", "application/vnd.amazon.eventstream")
		enc := eventstream.NewEncoder()
		writeConverseEvent(t, w, enc, "messageStart", `{"role":"assistant"}`)
		writeConverseEvent(t, w, enc, "contentBlockDelta",
			`{"contentBlockIndex":0,"delta":{"text":"half an ans"}}`)
		writeConverseEvent(t, w, enc, "contentBlockStop", `{"contentBlockIndex":0}`)
		writeConverseEvent(t, w, enc, "messageStop", `{"stopReason":"`+stopReason+`"}`)
		writeConverseEvent(t, w, enc, "metadata",
			`{"usage":{"inputTokens":1,"outputTokens":1,"totalTokens":2}}`)
	}))
	t.Cleanup(srv.Close)

	return bedrockLLMAgainst(t, srv,
		bedrock.WithModel("anthropic.claude-sonnet-4-5-20250929-v1:0"), bedrock.WithConverseAPI())
}

func streamedTurn(t *testing.T, stopReason string, extra ...llms.CallOption) (*llms.ContentResponse, error) {
	t.Helper()

	opts := append([]llms.CallOption{
		llms.WithStreamingFunc(func(_ context.Context, _ streaming.Chunk) error { return nil }),
	}, extra...)

	return converseStreamLLM(t, stopReason).GenerateContent(context.Background(),
		[]llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "hi")}, opts...)
}

func TestConverseStreamReportsTruncation(t *testing.T) {
	t.Parallel()

	t.Run("the stop reason reaches the choice", func(t *testing.T) {
		t.Parallel()
		got, err := streamedTurn(t, "max_tokens")
		require.NoError(t, err)

		choice := got.Choices[0]
		assert.Equal(t, "max_tokens", choice.StopReason)
		assert.True(t, choice.Truncated, "a max_tokens stop marks the choice truncated")
		assert.Equal(t, "half an ans", choice.Content)
	})

	t.Run("finishing on its own is not truncation", func(t *testing.T) {
		t.Parallel()
		got, err := streamedTurn(t, "end_turn")
		require.NoError(t, err)

		assert.Equal(t, "end_turn", got.Choices[0].StopReason)
		assert.False(t, got.Choices[0].Truncated)
	})

	t.Run("failing on truncation reaches the streamed answer too", func(t *testing.T) {
		t.Parallel()
		got, err := streamedTurn(t, "max_tokens", llms.WithFailOnTruncation())
		require.Error(t, err)
		assert.True(t, llms.IsTruncatedError(err))
		require.NotNil(t, got)
		assert.Equal(t, "half an ans", got.Choices[0].Content)
	})
}
