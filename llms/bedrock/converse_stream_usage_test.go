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

func TestConverseStreamReportsTheSameCountersAsTheWholeAnswer(t *testing.T) {
	t.Parallel()

	const usage = `{"usage":{"inputTokens":13,"outputTokens":304,"totalTokens":3919,` +
		`"cacheReadInputTokens":3602,"cacheWriteInputTokens":300}}`

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, r.Body)
		w.Header().Set("Content-Type", "application/vnd.amazon.eventstream")
		enc := eventstream.NewEncoder()
		writeConverseEvent(t, w, enc, "messageStart", `{"role":"assistant"}`)
		writeConverseEvent(t, w, enc, "contentBlockDelta",
			`{"contentBlockIndex":0,"delta":{"text":"ok"}}`)
		writeConverseEvent(t, w, enc, "contentBlockStop", `{"contentBlockIndex":0}`)
		writeConverseEvent(t, w, enc, "messageStop", `{"stopReason":"end_turn"}`)
		writeConverseEvent(t, w, enc, "metadata", usage)
	}))
	t.Cleanup(srv.Close)

	llm := bedrockLLMAgainst(t, srv,
		bedrock.WithModel("anthropic.claude-sonnet-4-5-20250929-v1:0"), bedrock.WithConverseAPI())

	resp, err := llm.GenerateContent(context.Background(),
		[]llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "hi")},
		llms.WithStreamingFunc(func(_ context.Context, _ streaming.Chunk) error { return nil }))
	require.NoError(t, err)
	require.Len(t, resp.Choices, 1)

	info := resp.Choices[0].GenerationInfo
	assert.Equal(t, int32(304), info["CompletionTokens"])
	assert.Equal(t, int32(3919), info["TotalTokens"])
	assert.Equal(t, int32(3602), info["CacheReadInputTokens"])
	assert.Equal(t, int32(3602), info["PromptCachedTokens"])
	assert.Equal(t, int32(300), info["CacheCreationInputTokens"])
	assert.Equal(t, int32(3915), info["PromptTokens"])
}
