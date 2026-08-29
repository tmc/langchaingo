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

func legacyStreamOfThreeWords(t *testing.T) *httptest.Server {
	t.Helper()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, r.Body)
		w.Header().Set("Content-Type", "application/vnd.amazon.eventstream")
		enc := eventstream.NewEncoder()
		writeLegacyChunk(t, w, enc, `{"type":"message_start","message":{"id":"x","type":"message",`+
			`"role":"assistant","model":"m","content":[],"stop_reason":null,`+
			`"usage":{"input_tokens":10,"output_tokens":1}}}`)
		for _, text := range []string{"sixty ", "rooms ", "are free"} {
			writeLegacyChunk(t, w, enc, `{"type":"content_block_delta","index":0,`+
				`"delta":{"type":"text_delta","text":"`+text+`"}}`)
		}
		writeLegacyChunk(t, w, enc, `{"type":"message_delta","delta":{"stop_reason":"end_turn"},`+
			`"usage":{"output_tokens":9}}`)
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestLegacyStreamKeepsWhatArrivedWhenTheConsumerGivesUp(t *testing.T) {
	t.Parallel()

	llm := bedrockLLMAgainst(t, legacyStreamOfThreeWords(t),
		bedrock.WithModel("anthropic.claude-sonnet-4-5-20250929-v1:0"))

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
		"the chunk the consumer refused was still handed to it, so it belongs in the answer")
}
