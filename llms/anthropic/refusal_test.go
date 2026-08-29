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
