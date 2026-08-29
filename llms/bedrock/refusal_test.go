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

func TestLegacyRefusalReachesTheCallerAsATypedError(t *testing.T) {
	t.Parallel()

	const resp = `{"id":"x","type":"message","role":"assistant","model":"m",` +
		`"content":[{"type":"text","text":"I can't help with that."}],` +
		`"stop_reason":"refusal",` +
		`"stop_details":{"category":"cyber","explanation":"content boundary"},` +
		`"usage":{"input_tokens":11,"output_tokens":7,"cache_creation_input_tokens":3,"cache_read_input_tokens":5}}`

	llm, _ := legacyLLMCapturing(t, resp, bedrock.WithModel(claudeBudgetModel))
	got, err := llm.GenerateContent(context.Background(),
		[]llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "hi")})

	var refusal *llms.ErrModelRefusal
	require.ErrorAs(t, err, &refusal)
	assert.Equal(t, "I can't help with that.", refusal.Message)
	assert.Equal(t, "cyber", refusal.Category, "stop_details must survive the door")
	assert.Equal(t, "content boundary", refusal.Explanation)
	assert.Equal(t, 11, refusal.InputTokens)
	assert.Equal(t, 7, refusal.OutputTokens)
	assert.Equal(t, 3, refusal.CacheCreationInputTokens)
	assert.Equal(t, 5, refusal.CacheReadInputTokens)

	require.NotNil(t, got, "the response travels with the error so usage survives")
}

func TestLegacyStreamRefusalIsNotMistakenForAnAnswer(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, r.Body)
		w.Header().Set("Content-Type", "application/vnd.amazon.eventstream")
		enc := eventstream.NewEncoder()

		writeLegacyChunk(t, w, enc, `{"type":"message_start","message":{"id":"x","type":"message",`+
			`"role":"assistant","model":"m","content":[],"stop_reason":null,`+
			`"usage":{"input_tokens":11,"output_tokens":1}}}`)
		writeLegacyChunk(t, w, enc, `{"type":"content_block_delta","index":0,`+
			`"delta":{"type":"text_delta","text":"I can't help with that."}}`)
		writeLegacyChunk(t, w, enc, `{"type":"message_delta","delta":{"stop_reason":"refusal"},`+
			`"usage":{"output_tokens":7}}`)
	}))
	t.Cleanup(srv.Close)

	llm := bedrockLLMAgainst(t, srv, bedrock.WithModel("anthropic.claude-sonnet-4-5-20250929-v1:0"))

	got, err := llm.GenerateContent(context.Background(),
		[]llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "hi")},
		llms.WithStreamingFunc(func(context.Context, streaming.Chunk) error { return nil }))

	var refusal *llms.ErrModelRefusal
	require.ErrorAs(t, err, &refusal, "a streamed refusal must not look like a normal answer")
	assert.Equal(t, "I can't help with that.", refusal.Message)
	assert.Equal(t, 11, refusal.InputTokens)
	assert.Equal(t, 7, refusal.OutputTokens)

	require.NotNil(t, got, "the response travels with the error so usage survives")
	require.Len(t, got.Choices, 1)
	assert.Equal(t, "refusal", got.Choices[0].StopReason)
}
