package anthropic

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vxcontrol/langchaingo/llms"
)

// serveAnthropic's returned pointer holds the request body only once
// GenerateContent has returned.
func serveAnthropic(t *testing.T, response string) (*httptest.Server, *string) {
	t.Helper()

	sent := new(string)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		*sent = string(b)
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, response)
	}))
	t.Cleanup(srv.Close)

	return srv, sent
}

func TestEveryTextBlockOfATurnReachesTheChoice(t *testing.T) {
	t.Parallel()

	const resp = `{"id":"x","type":"message","role":"assistant","model":"m","content":[` +
		`{"type":"text","text":"17 * 23 = "},` +
		`{"type":"tool_use","id":"t1","name":"calc","input":{"expr":"17*23"}},` +
		`{"type":"text","text":"391."}],` +
		`"stop_reason":"end_turn","usage":{"input_tokens":1,"output_tokens":1}}`

	srv, _ := serveAnthropic(t, resp)

	llm, err := New(WithBaseURL(srv.URL), WithToken("t"), WithModel("claude-sonnet-4-5-20250929"))
	require.NoError(t, err)

	got, err := llm.GenerateContent(context.Background(),
		[]llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "hi")})
	require.NoError(t, err)

	assert.Equal(t, "17 * 23 = 391.", got.Choices[0].Content,
		"a turn that resumes text after a tool call must keep both segments")
	assert.Len(t, got.Choices[0].ToolCalls, 1)
}

func TestAnEffortWithNoBudgetSendsNoThinking(t *testing.T) {
	t.Parallel()

	const resp = `{"id":"x","type":"message","role":"assistant","model":"m",` +
		`"content":[{"type":"text","text":"ok"}],"stop_reason":"end_turn",` +
		`"usage":{"input_tokens":1,"output_tokens":1}}`

	srv, sent := serveAnthropic(t, resp)

	llm, err := New(WithBaseURL(srv.URL), WithToken("t"), WithModel("claude-sonnet-4-5-20250929"))
	require.NoError(t, err)

	_, err = llm.GenerateContent(context.Background(),
		[]llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "hi")},
		llms.WithReasoning(llms.ReasoningEffort("minimal"), 0))
	require.NoError(t, err)

	assert.NotContains(t, *sent, `"budget_tokens":-1`,
		"an effort the budget table does not map must not travel as a negative budget")
	assert.NotContains(t, *sent, `"thinking"`,
		"no budget means no thinking payload, as on both Bedrock doors")
}

func TestABudgetedEffortStillSendsThinking(t *testing.T) {
	t.Parallel()

	const resp = `{"id":"x","type":"message","role":"assistant","model":"m",` +
		`"content":[{"type":"text","text":"ok"}],"stop_reason":"end_turn",` +
		`"usage":{"input_tokens":1,"output_tokens":1}}`

	srv, sent := serveAnthropic(t, resp)

	llm, err := New(WithBaseURL(srv.URL), WithToken("t"), WithModel("claude-sonnet-4-5-20250929"))
	require.NoError(t, err)

	_, err = llm.GenerateContent(context.Background(),
		[]llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "hi")},
		llms.WithReasoning(llms.ReasoningHigh, 0), llms.WithMaxTokens(8000))
	require.NoError(t, err)

	assert.Contains(t, *sent, `"type":"enabled"`)
	require.False(t, strings.Contains(*sent, `"budget_tokens":0`),
		"a mapped effort must carry a positive budget")
}
