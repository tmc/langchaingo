package openai

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/llms/reasoning"
)

func turnLimitErr(t *testing.T, model string, msgs []llms.MessageContent, opts ...llms.CallOption) error {
	t.Helper()

	const completion = `{"id":"x","object":"chat.completion","created":1,"model":"m",` +
		`"choices":[{"index":0,"message":{"role":"assistant","content":"ok"},"finish_reason":"stop"}],` +
		`"usage":{"prompt_tokens":1,"completion_tokens":1,"total_tokens":2}}`

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, r.Body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, completion)
	}))
	t.Cleanup(srv.Close)

	llm, err := New(WithBaseURL(srv.URL), WithToken("test"), WithModel(model))
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	_, err = llm.GenerateContent(context.Background(), msgs, opts...)
	return err
}

func askedFor(text string) []llms.MessageContent {
	return []llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, text)}
}

func endingOnAssistant() []llms.MessageContent {
	return []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeHuman, "finish this"),
		llms.TextParts(llms.ChatMessageTypeAI, "the answer is"),
	}
}

func TestClaudeTurnLimitsOnTheOpenAITransport(t *testing.T) {
	t.Parallel()

	forced := llms.WithToolChoice(llms.ToolChoice{Type: "any"})
	thinking := llms.WithReasoning(llms.ReasoningLow, 0)

	t.Run("budget thinking with a forced tool is refused", func(t *testing.T) {
		t.Parallel()
		err := turnLimitErr(t, "claude-sonnet-4-5", askedFor("hi"), thinking, forced)
		var target *reasoning.ErrForcedToolUseWithThinking
		if !errors.As(err, &target) {
			t.Errorf("want ErrForcedToolUseWithThinking, got %v", err)
		}
	})

	t.Run("adaptive thinking carries no such limit", func(t *testing.T) {
		t.Parallel()
		if err := turnLimitErr(t, "claude-sonnet-5", askedFor("hi"), thinking, forced); err != nil {
			t.Errorf("an adaptive generation takes a forced tool alongside thinking, got %v", err)
		}
	})

	t.Run("a non-Claude model is not refused", func(t *testing.T) {
		t.Parallel()
		if err := turnLimitErr(t, "gpt-5.2", askedFor("hi"), thinking, forced); err != nil {
			t.Errorf("the rule is Anthropic's, got %v", err)
		}
	})

	t.Run("a generation that rejects prefill is refused", func(t *testing.T) {
		t.Parallel()
		err := turnLimitErr(t, "claude-opus-4-6", endingOnAssistant())
		var target *reasoning.ErrAssistantPrefillUnsupported
		if !errors.As(err, &target) {
			t.Errorf("want ErrAssistantPrefillUnsupported, got %v", err)
		}
	})

	t.Run("an older generation still takes a prefill", func(t *testing.T) {
		t.Parallel()
		if err := turnLimitErr(t, "claude-sonnet-4-5", endingOnAssistant()); err != nil {
			t.Errorf("this generation answers a prefilled turn, got %v", err)
		}
	})
}
