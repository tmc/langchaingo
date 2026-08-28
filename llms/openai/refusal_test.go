package openai

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/vxcontrol/langchaingo/llms"
)

func TestARefusalReachesTheCallerAsATypedError(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, r.Body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"id":"x","object":"chat.completion","created":1,"model":"gpt-4o",
			"choices":[{"index":0,"message":{"role":"assistant","content":"",
			"refusal":"I can't help with that."},"finish_reason":"stop"}],
			"usage":{"prompt_tokens":11,"completion_tokens":7,"total_tokens":18}}`)
	}))
	t.Cleanup(srv.Close)

	llm := newUnitLLM(t, WithBaseURL(srv.URL), WithModel("gpt-4o"))
	resp, err := llm.GenerateContent(context.Background(),
		[]llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "hi")})

	var refusal *llms.ErrModelRefusal
	require.ErrorAs(t, err, &refusal)
	assert.Equal(t, "I can't help with that.", refusal.Message)
	assert.Equal(t, 11, refusal.InputTokens)
	assert.Equal(t, 7, refusal.OutputTokens)

	require.NotNil(t, resp, "the response travels with the error so usage survives")
	assert.Equal(t, "I can't help with that.", resp.Choices[0].GenerationInfo["Refusal"])
}
