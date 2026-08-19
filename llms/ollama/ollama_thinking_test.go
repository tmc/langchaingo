package ollama

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/llms/streaming"
)

func newThinkingServerClient(t *testing.T, lines ...string) *LLM {
	t.Helper()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/x-ndjson")
		for _, line := range lines {
			_, _ = w.Write([]byte(line + "\n"))
		}
	}))
	t.Cleanup(srv.Close)

	llm, err := New(WithServerURL(srv.URL), WithModel("glm-5"))
	require.NoError(t, err)
	return llm
}

func TestNativeThinkingReachesTheChoice(t *testing.T) {
	t.Parallel()

	llm := newThinkingServerClient(t,
		`{"model":"glm-5","message":{"role":"assistant","content":"Paris",`+
			`"thinking":"The capital of France is Paris."},"done":true,"done_reason":"stop"}`)

	resp, err := llm.GenerateContent(t.Context(),
		[]llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "capital of France?")})
	require.NoError(t, err)
	require.Len(t, resp.Choices, 1)

	require.Equal(t, "Paris", resp.Choices[0].Content)
	require.NotNil(t, resp.Choices[0].Reasoning)
	require.Equal(t, "The capital of France is Paris.", resp.Choices[0].Reasoning.Content)
}

func TestNativeThinkingIsStreamedAndAccumulated(t *testing.T) {
	t.Parallel()

	llm := newThinkingServerClient(t,
		`{"model":"glm-5","message":{"role":"assistant","content":"","thinking":"first "},"done":false}`,
		`{"model":"glm-5","message":{"role":"assistant","content":"","thinking":"second"},"done":false}`,
		`{"model":"glm-5","message":{"role":"assistant","content":"Paris"},"done":true,"done_reason":"stop"}`)

	var streamed string
	resp, err := llm.GenerateContent(t.Context(),
		[]llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "capital of France?")},
		llms.WithStreamingFunc(func(_ context.Context, chunk streaming.Chunk) error {
			if chunk.Type == streaming.ChunkTypeReasoning && chunk.Reasoning != nil {
				streamed += chunk.Reasoning.Content
			}
			return nil
		}))
	require.NoError(t, err)

	require.Equal(t, "first second", streamed, "every thinking delta must reach the callback")
	require.NotNil(t, resp.Choices[0].Reasoning)
	require.Equal(t, "first second", resp.Choices[0].Reasoning.Content)
}

func TestLegacyThinkTagsStillSplit(t *testing.T) {
	t.Parallel()

	body, err := json.Marshal(map[string]any{
		"model":       "deepseek-r1",
		"message":     map[string]any{"role": "assistant", "content": "<think>musing</think>Paris"},
		"done":        true,
		"done_reason": "stop",
	})
	require.NoError(t, err)

	llm := newThinkingServerClient(t, string(body))
	resp, err := llm.GenerateContent(t.Context(),
		[]llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "capital of France?")})
	require.NoError(t, err)

	require.Equal(t, "Paris", resp.Choices[0].Content)
	require.Equal(t, "musing", resp.Choices[0].Reasoning.Content)
}
