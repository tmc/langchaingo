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

func TestConverseStreamKeepsToolCallsApart(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, r.Body)
		w.Header().Set("Content-Type", "application/vnd.amazon.eventstream")
		enc := eventstream.NewEncoder()
		writeConverseEvent(t, w, enc, "messageStart", `{"role":"assistant"}`)

		writeConverseEvent(t, w, enc, "contentBlockStart",
			`{"contentBlockIndex":0,"start":{"toolUse":{"toolUseId":"id-a","name":"alpha"}}}`)
		writeConverseEvent(t, w, enc, "contentBlockStart",
			`{"contentBlockIndex":1,"start":{"toolUse":{"toolUseId":"id-b","name":"beta"}}}`)

		writeConverseEvent(t, w, enc, "contentBlockDelta",
			`{"contentBlockIndex":0,"delta":{"toolUse":{"input":"{\"a\":"}}}`)
		writeConverseEvent(t, w, enc, "contentBlockDelta",
			`{"contentBlockIndex":1,"delta":{"toolUse":{"input":"{\"b\":"}}}`)
		writeConverseEvent(t, w, enc, "contentBlockDelta",
			`{"contentBlockIndex":0,"delta":{"toolUse":{"input":"1}"}}}`)
		writeConverseEvent(t, w, enc, "contentBlockDelta",
			`{"contentBlockIndex":1,"delta":{"toolUse":{"input":"2}"}}}`)

		writeConverseEvent(t, w, enc, "contentBlockStop", `{"contentBlockIndex":0}`)
		writeConverseEvent(t, w, enc, "contentBlockStop", `{"contentBlockIndex":1}`)
		writeConverseEvent(t, w, enc, "messageStop", `{"stopReason":"tool_use"}`)
		writeConverseEvent(t, w, enc, "metadata",
			`{"usage":{"inputTokens":1,"outputTokens":1,"totalTokens":2}}`)
	}))
	t.Cleanup(srv.Close)

	llm := bedrockLLMAgainst(t, srv,
		bedrock.WithModel("anthropic.claude-sonnet-4-5-20250929-v1:0"), bedrock.WithConverseAPI())

	var streamed []streaming.ToolCall
	resp, err := llm.GenerateContent(context.Background(),
		[]llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "hi")},
		llms.WithStreamingFunc(func(_ context.Context, chunk streaming.Chunk) error {
			if chunk.Type == streaming.ChunkTypeToolCall {
				streamed = append(streamed, chunk.ToolCall)
			}
			return nil
		}))
	require.NoError(t, err)

	got := map[string]string{}
	for _, call := range resp.Choices[0].ToolCalls {
		require.NotNil(t, call.FunctionCall)
		got[call.FunctionCall.Name] = call.FunctionCall.Arguments
	}

	assert.Len(t, resp.Choices[0].ToolCalls, 2)
	assert.Equal(t, `{"a":1}`, got["alpha"])
	assert.Equal(t, `{"b":2}`, got["beta"])
	assert.Len(t, streamed, 2)
}

func TestConverseStreamToolCallWithoutABlockIndex(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, r.Body)
		w.Header().Set("Content-Type", "application/vnd.amazon.eventstream")
		enc := eventstream.NewEncoder()
		writeConverseEvent(t, w, enc, "messageStart", `{"role":"assistant"}`)
		writeConverseEvent(t, w, enc, "contentBlockStart",
			`{"start":{"toolUse":{"toolUseId":"id-a","name":"alpha"}}}`)
		writeConverseEvent(t, w, enc, "contentBlockDelta",
			`{"delta":{"toolUse":{"input":"{\"a\":1}"}}}`)
		writeConverseEvent(t, w, enc, "contentBlockStop", `{}`)
		writeConverseEvent(t, w, enc, "messageStop", `{"stopReason":"tool_use"}`)
		writeConverseEvent(t, w, enc, "metadata",
			`{"usage":{"inputTokens":1,"outputTokens":1,"totalTokens":2}}`)
	}))
	t.Cleanup(srv.Close)

	llm := bedrockLLMAgainst(t, srv,
		bedrock.WithModel("anthropic.claude-sonnet-4-5-20250929-v1:0"), bedrock.WithConverseAPI())

	resp, err := llm.GenerateContent(context.Background(),
		[]llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "hi")},
		llms.WithStreamingFunc(func(_ context.Context, _ streaming.Chunk) error { return nil }))
	require.NoError(t, err)

	require.Len(t, resp.Choices[0].ToolCalls, 1)
	require.NotNil(t, resp.Choices[0].ToolCalls[0].FunctionCall)
	assert.Equal(t, `{"a":1}`, resp.Choices[0].ToolCalls[0].FunctionCall.Arguments)
}
