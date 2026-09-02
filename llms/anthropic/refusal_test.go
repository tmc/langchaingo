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

func TestARefusalTravelsWithTheResponseOnThePrimaryDoor(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, r.Body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"id":"msg_x","type":"message","role":"assistant","model":"claude-opus-4-6",
			"content":[{"type":"text","text":"I can't help with that."}],"stop_reason":"refusal",
			"stop_details":{"category":"policy","explanation":"declined by policy"},
			"usage":{"input_tokens":11,"output_tokens":7}}`)
	}))
	t.Cleanup(srv.Close)

	llm, err := anthropic.New(anthropic.WithToken("test-key"), anthropic.WithBaseURL(srv.URL),
		anthropic.WithModel("claude-opus-4-6"))
	require.NoError(t, err)

	resp, err := llm.GenerateContent(context.Background(),
		[]llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "hi")})

	var refusal *llms.ErrModelRefusal
	require.ErrorAs(t, err, &refusal)
	assert.Equal(t, "I can't help with that.", refusal.Message)
	assert.Equal(t, 11, refusal.InputTokens)
	assert.Equal(t, 7, refusal.OutputTokens)

	require.NotNil(t, resp, "the response travels with the error so usage survives")
	require.NotEmpty(t, resp.Choices)
	assert.Equal(t, "refusal", resp.Choices[0].StopReason)
}

func TestAStreamedRefusalAlsoTravelsWithItsResponse(t *testing.T) {
	t.Parallel()

	const events = `event: message_start
data: {"type":"message_start","message":{"id":"msg_1","type":"message","role":"assistant","model":"claude-haiku-4-5","content":[],"stop_reason":null,"usage":{"input_tokens":11,"output_tokens":1}}}

event: content_block_start
data: {"type":"content_block_start","index":0,"content_block":{"type":"text","text":""}}

event: content_block_delta
data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"I can't help with that."}}

event: content_block_stop
data: {"type":"content_block_stop","index":0}

event: message_delta
data: {"type":"message_delta","delta":{"stop_reason":"refusal","stop_sequence":null},"usage":{"output_tokens":7}}

event: message_stop
data: {"type":"message_stop"}
`

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, r.Body)
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = io.WriteString(w, events)
	}))
	t.Cleanup(srv.Close)

	llm, err := anthropic.New(anthropic.WithToken("test-key"), anthropic.WithBaseURL(srv.URL),
		anthropic.WithModel("claude-haiku-4-5"))
	require.NoError(t, err)

	resp, err := llm.GenerateContent(context.Background(),
		[]llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "hi")},
		llms.WithStreamingFunc(func(context.Context, streaming.Chunk) error { return nil }))

	var refusal *llms.ErrModelRefusal
	require.ErrorAs(t, err, &refusal)
	assert.Equal(t, "I can't help with that.", refusal.Message)

	require.NotNil(t, resp, "the streamed door must answer a refusal the same way")
	require.NotEmpty(t, resp.Choices)
	assert.Equal(t, "refusal", resp.Choices[0].StopReason)
}

func TestARefusalKeepsTheCacheBreakdownAndTheServiceTier(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, r.Body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"id":"msg_x","type":"message","role":"assistant","model":"claude-opus-4-6",
			"content":[{"type":"text","text":"I can't help with that."}],"stop_reason":"refusal",
			"usage":{"input_tokens":11,"output_tokens":7,"cache_read_input_tokens":3,
			"cache_creation_input_tokens":9,
			"cache_creation":{"ephemeral_5m_input_tokens":4,"ephemeral_1h_input_tokens":5},
			"service_tier":"priority","speed":"fast"}}`)
	}))
	t.Cleanup(srv.Close)

	llm, err := anthropic.New(anthropic.WithToken("test-key"), anthropic.WithBaseURL(srv.URL),
		anthropic.WithModel("claude-opus-4-6"))
	require.NoError(t, err)

	resp, err := llm.GenerateContent(context.Background(),
		[]llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "hi")})

	var refusal *llms.ErrModelRefusal
	require.ErrorAs(t, err, &refusal)
	assert.Equal(t, 9, refusal.CacheCreationInputTokens)
	assert.Equal(t, 3, refusal.CacheReadInputTokens)

	require.NotNil(t, resp)
	require.NotEmpty(t, resp.Choices)
	info := resp.Choices[0].GenerationInfo
	assert.Equal(t, 4, info["CacheCreationEphemeral5mInputTokens"])
	assert.Equal(t, 5, info["CacheCreationEphemeral1hInputTokens"])
	assert.Equal(t, "priority", info["ServiceTier"])
	assert.Equal(t, "fast", info["InferenceSpeed"])
}
