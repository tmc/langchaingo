package anthropic_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/llms/anthropic"
	"github.com/vxcontrol/langchaingo/llms/streaming"
)

func answeringLLM(t *testing.T, contentType, body string) *anthropic.LLM {
	t.Helper()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, r.Body)
		w.Header().Set("Content-Type", contentType)
		_, _ = io.WriteString(w, body)
	}))
	t.Cleanup(srv.Close)

	llm, err := anthropic.New(
		anthropic.WithToken("test-key"),
		anthropic.WithBaseURL(srv.URL),
		anthropic.WithModel("claude-sonnet-4-5"))
	require.NoError(t, err)
	return llm
}

func TestReasoningTokensReachTheCaller(t *testing.T) {
	t.Parallel()

	ask := []llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "hi")}

	t.Run("the whole answer carries the count", func(t *testing.T) {
		t.Parallel()

		llm := answeringLLM(t, "application/json", `{"id":"msg_1","type":"message","role":"assistant",
			"model":"claude-sonnet-4-5",
			"content":[{"type":"thinking","thinking":"counting","signature":"c2ln"},
				{"type":"text","text":"391"}],
			"stop_reason":"end_turn",
			"usage":{"input_tokens":55,"output_tokens":352,
				"output_tokens_details":{"thinking_tokens":169}}}`)

		resp, err := llm.GenerateContent(context.Background(), ask, llms.WithMaxTokens(2048))
		require.NoError(t, err)

		assert.Equal(t, 169, resp.Choices[0].GenerationInfo["ReasoningTokens"])
		assert.Equal(t, 352, resp.Choices[0].GenerationInfo["CompletionTokens"])
	})

	t.Run("a streamed answer carries the same count", func(t *testing.T) {
		t.Parallel()

		llm := answeringLLM(t, "text/event-stream", `event: message_start
data: {"type":"message_start","message":{"id":"msg_2","type":"message","role":"assistant","model":"claude-sonnet-4-5","content":[],"stop_reason":null,"usage":{"input_tokens":55,"output_tokens":1}}}

event: content_block_start
data: {"type":"content_block_start","index":0,"content_block":{"type":"text","text":""}}

event: content_block_delta
data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"391"}}

event: content_block_stop
data: {"type":"content_block_stop","index":0}

event: message_delta
data: {"type":"message_delta","delta":{"stop_reason":"end_turn","stop_sequence":null},"usage":{"output_tokens":352,"output_tokens_details":{"thinking_tokens":169}}}

event: message_stop
data: {"type":"message_stop"}
`)

		resp, err := llm.GenerateContent(context.Background(), ask,
			llms.WithMaxTokens(2048),
			llms.WithStreamingFunc(func(_ context.Context, _ streaming.Chunk) error { return nil }))
		require.NoError(t, err)

		assert.Equal(t, 169, resp.Choices[0].GenerationInfo["ReasoningTokens"])
	})

	t.Run("an answer without thinking reports none", func(t *testing.T) {
		t.Parallel()

		llm := answeringLLM(t, "application/json", `{"id":"msg_3","type":"message","role":"assistant",
			"model":"claude-sonnet-4-5","content":[{"type":"text","text":"ready"}],
			"stop_reason":"end_turn",
			"usage":{"input_tokens":16,"output_tokens":4,
				"output_tokens_details":{"thinking_tokens":0}}}`)

		resp, err := llm.GenerateContent(context.Background(), ask, llms.WithMaxTokens(2048))
		require.NoError(t, err)

		assert.Equal(t, 0, resp.Choices[0].GenerationInfo["ReasoningTokens"])
	})
}
