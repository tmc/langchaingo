package anthropic

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/schema"
)

// callbackRecorder records GenerateContent callback invocations so tests can
// assert that start/end/error hooks fire exactly once and on the right paths.
type callbackRecorder struct {
	starts  int
	ends    int
	errors  int
	lastEnd *llms.ContentResponse
}

func (r *callbackRecorder) HandleText(ctx context.Context, text string) {}

func (r *callbackRecorder) HandleLLMStart(ctx context.Context, prompts []string) {}

func (r *callbackRecorder) HandleLLMGenerateContentStart(ctx context.Context, ms []llms.MessageContent) {
	r.starts++
}

func (r *callbackRecorder) HandleLLMGenerateContentEnd(ctx context.Context, res *llms.ContentResponse) {
	r.ends++
	r.lastEnd = res
}

func (r *callbackRecorder) HandleLLMError(ctx context.Context, err error) {
	r.errors++
}

func (r *callbackRecorder) HandleChainStart(ctx context.Context, inputs map[string]any) {}

func (r *callbackRecorder) HandleChainEnd(ctx context.Context, outputs map[string]any) {}

func (r *callbackRecorder) HandleChainError(ctx context.Context, err error) {}

func (r *callbackRecorder) HandleToolStart(ctx context.Context, input string) {}

func (r *callbackRecorder) HandleToolEnd(ctx context.Context, output string) {}

func (r *callbackRecorder) HandleToolError(ctx context.Context, err error) {}

func (r *callbackRecorder) HandleAgentAction(ctx context.Context, action schema.AgentAction) {}

func (r *callbackRecorder) HandleAgentFinish(ctx context.Context, finish schema.AgentFinish) {}

func (r *callbackRecorder) HandleRetrieverStart(ctx context.Context, query string) {}

func (r *callbackRecorder) HandleRetrieverEnd(ctx context.Context, query string, documents []schema.Document) {
}

func (r *callbackRecorder) HandleStreamingFunc(ctx context.Context, chunk []byte) {}

// messagesResponseJSON is a minimal valid Messages API response.
const messagesResponseJSON = `{
	"id": "msg_test",
	"type": "message",
	"role": "assistant",
	"model": "claude-haiku-4-5",
	"content": [{"type": "text", "text": "hello"}],
	"stop_reason": "end_turn",
	"usage": {"input_tokens": 1, "output_tokens": 1}
}`

// completionResponseJSON is a minimal valid legacy text completions response.
const completionResponseJSON = `{
	"completion": "hello",
	"id": "cmpl_test",
	"model": "claude-haiku-4-5",
	"stop_reason": "end_turn"
}`

// newCallbackTestLLM spins up a local Anthropic API stub and wires an LLM
// against it, attaching the given callback recorder.
func newCallbackTestLLM(t *testing.T, handler *callbackRecorder, opts ...Option) *LLM {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.HasSuffix(r.URL.Path, "/complete"):
			_, _ = w.Write([]byte(completionResponseJSON))
		case strings.HasSuffix(r.URL.Path, "/messages"):
			_, _ = w.Write([]byte(messagesResponseJSON))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(server.Close)

	llm, err := New(append([]Option{
		WithToken("test-token"),
		WithBaseURL(server.URL),
		WithHTTPClient(server.Client()),
	}, opts...)...)
	require.NoError(t, err)
	llm.CallbacksHandler = handler

	return llm
}

func TestGenerateContentCallbacks(t *testing.T) {
	userMessages := []llms.MessageContent{
		{Role: llms.ChatMessageTypeHuman, Parts: []llms.ContentPart{llms.TextContent{Text: "hi"}}},
	}

	t.Run("messages_path_success_calls_end", func(t *testing.T) {
		rec := &callbackRecorder{}
		llm := newCallbackTestLLM(t, rec)

		resp, err := llm.GenerateContent(context.Background(), userMessages)
		require.NoError(t, err)
		require.NotNil(t, resp)

		assert.Equal(t, 1, rec.starts)
		assert.Equal(t, 1, rec.ends, "successful GenerateContent must call HandleLLMGenerateContentEnd")
		assert.Equal(t, 0, rec.errors)
		assert.Same(t, resp, rec.lastEnd)
	})

	t.Run("legacy_completions_path_success_calls_end", func(t *testing.T) {
		rec := &callbackRecorder{}
		llm := newCallbackTestLLM(t, rec, WithLegacyTextCompletionsAPI())

		resp, err := llm.GenerateContent(context.Background(), userMessages)
		require.NoError(t, err)
		require.NotNil(t, resp)

		assert.Equal(t, 1, rec.starts)
		assert.Equal(t, 1, rec.ends, "successful GenerateContent must call HandleLLMGenerateContentEnd")
		assert.Equal(t, 0, rec.errors)
		assert.Same(t, resp, rec.lastEnd)
	})

	t.Run("error_path_calls_error_not_end", func(t *testing.T) {
		rec := &callbackRecorder{}
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		t.Cleanup(server.Close)

		llm, err := New(WithToken("test-token"), WithBaseURL(server.URL), WithHTTPClient(server.Client()))
		require.NoError(t, err)
		llm.CallbacksHandler = rec

		_, err = llm.GenerateContent(context.Background(), userMessages)
		require.Error(t, err)

		assert.Equal(t, 1, rec.starts)
		assert.Equal(t, 1, rec.errors)
		assert.Equal(t, 0, rec.ends, "failed GenerateContent must not call HandleLLMGenerateContentEnd")
	})
}
