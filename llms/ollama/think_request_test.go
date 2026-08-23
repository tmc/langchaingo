package ollama

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vxcontrol/langchaingo/llms"
)

func captureChatRequest(t *testing.T, opts ...llms.CallOption) map[string]any {
	t.Helper()

	var body []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/x-ndjson")
		_, _ = w.Write([]byte(`{"model":"glm-5","message":{"role":"assistant","content":"ok"},` +
			`"done":true,"done_reason":"stop"}` + "\n"))
	}))
	t.Cleanup(srv.Close)

	llm, err := New(WithServerURL(srv.URL), WithModel("glm-5"))
	require.NoError(t, err)

	_, err = llm.GenerateContent(t.Context(),
		[]llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "hi")}, opts...)
	require.NoError(t, err)

	var got map[string]any
	require.NoError(t, json.Unmarshal(body, &got))
	return got
}

func TestThinkReachesTheWire(t *testing.T) {
	t.Parallel()

	t.Run("disabled reasoning is sent, not left to the server", func(t *testing.T) {
		t.Parallel()
		require.Equal(t, false, captureChatRequest(t, llms.WithReasoningDisabled())["think"])
	})

	t.Run("a level ollama accepts travels as itself", func(t *testing.T) {
		t.Parallel()
		require.Equal(t, "low", captureChatRequest(t, llms.WithReasoning(llms.ReasoningLow, 0))["think"])
	})

	t.Run("a level ollama does not accept falls back to plain on", func(t *testing.T) {
		t.Parallel()
		require.Equal(t, true, captureChatRequest(t, llms.WithReasoning(llms.ReasoningXHigh, 0))["think"])
	})

	t.Run("saying nothing leaves the field off the wire", func(t *testing.T) {
		t.Parallel()
		require.NotContains(t, captureChatRequest(t), "think")
	})
}
