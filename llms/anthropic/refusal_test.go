package anthropic_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/llms/anthropic"
)

func generateAgainst(t *testing.T, body string) (*llms.ContentResponse, error) {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)

	llm, err := anthropic.New(anthropic.WithToken("k"), anthropic.WithBaseURL(srv.URL), anthropic.WithModel("claude-fable-5"))
	require.NoError(t, err)
	return llm.GenerateContent(t.Context(),
		[]llms.MessageContent{{Role: llms.ChatMessageTypeHuman, Parts: []llms.ContentPart{llms.TextPart("hi")}}})
}

// A refusal (stop_reason "refusal") must surface as a typed ErrModelRefusal
// carrying the refusal text, distinct from an empty/failed response.
func TestAnthropic_Refusal(t *testing.T) {
	t.Parallel()

	t.Run("refusal with text", func(t *testing.T) {
		t.Parallel()
		_, err := generateAgainst(t, `{"id":"m","type":"message","role":"assistant","model":"claude-fable-5",`+
			`"content":[{"type":"text","text":"I can't continue this story."}],"stop_reason":"refusal","usage":{"input_tokens":1,"output_tokens":1}}`)
		var refusal *anthropic.ErrModelRefusal
		require.ErrorAs(t, err, &refusal)
		assert.Equal(t, "I can't continue this story.", refusal.Message)
	})

	t.Run("empty refusal is a refusal, not an empty response", func(t *testing.T) {
		t.Parallel()
		_, err := generateAgainst(t, `{"id":"m","type":"message","role":"assistant","model":"claude-fable-5",`+
			`"content":[],"stop_reason":"refusal","usage":{"input_tokens":1,"output_tokens":1}}`)
		var refusal *anthropic.ErrModelRefusal
		require.ErrorAs(t, err, &refusal, "empty-content refusal must be ErrModelRefusal, not ErrEmptyResponse")
		assert.False(t, errors.Is(err, anthropic.ErrEmptyResponse))
	})

	t.Run("normal end_turn is unaffected", func(t *testing.T) {
		t.Parallel()
		resp, err := generateAgainst(t, `{"id":"m","type":"message","role":"assistant","model":"claude-fable-5",`+
			`"content":[{"type":"text","text":"ok"}],"stop_reason":"end_turn","usage":{"input_tokens":1,"output_tokens":1}}`)
		require.NoError(t, err)
		assert.Equal(t, "ok", resp.Choices[0].Content)
	})
}
