package bedrock_test

import (
	"context"
	"encoding/base64"
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

func TestLegacyStreamReportsTheSameCountersAsTheWholeAnswer(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, r.Body)
		w.Header().Set("Content-Type", "application/vnd.amazon.eventstream")
		enc := eventstream.NewEncoder()

		writeLegacyChunk(t, w, enc, `{"type":"message_start","message":{"id":"x","type":"message",`+
			`"role":"assistant","model":"m","content":[],"stop_reason":null,`+
			`"usage":{"input_tokens":13,"cache_creation_input_tokens":300,"cache_read_input_tokens":3602,`+
			`"output_tokens":1}}}`)
		writeLegacyChunk(t, w, enc, `{"type":"content_block_delta","index":0,`+
			`"delta":{"type":"text_delta","text":"ok"}}`)
		writeLegacyChunk(t, w, enc, `{"type":"message_delta","delta":{"stop_reason":"end_turn"},`+
			`"usage":{"output_tokens":304,"output_tokens_details":{"thinking_tokens":178}}}`)
	}))
	t.Cleanup(srv.Close)

	llm := bedrockLLMAgainst(t, srv, bedrock.WithModel("anthropic.claude-sonnet-4-5-20250929-v1:0"))

	got, err := llm.GenerateContent(context.Background(),
		[]llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "hi")},
		llms.WithStreamingFunc(func(context.Context, streaming.Chunk) error { return nil }))
	require.NoError(t, err)
	require.Len(t, got.Choices, 1)

	info := got.Choices[0].GenerationInfo
	assert.Equal(t, 13, info["input_tokens"])
	assert.Equal(t, 3915, info["PromptTokens"])
	assert.Equal(t, 304, info["CompletionTokens"])
	assert.Equal(t, 4219, info["TotalTokens"])
	assert.Equal(t, 3602, info["CacheReadInputTokens"])
	assert.Equal(t, 3602, info["PromptCachedTokens"])
	assert.Equal(t, 300, info["CacheCreationInputTokens"])
	assert.Equal(t, 178, info["ReasoningTokens"])
}

func writeLegacyChunk(t *testing.T, w io.Writer, enc *eventstream.Encoder, payload string) {
	t.Helper()

	inner := `{"bytes":"` + base64Encode(payload) + `"}`
	require.NoError(t, enc.Encode(w, eventstream.Message{
		Headers: eventstream.Headers{
			{Name: ":message-type", Value: eventstream.StringValue("event")},
			{Name: ":event-type", Value: eventstream.StringValue("chunk")},
			{Name: ":content-type", Value: eventstream.StringValue("application/json")},
		},
		Payload: []byte(inner),
	}))
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}
}

func base64Encode(s string) string { return base64.StdEncoding.EncodeToString([]byte(s)) }
