package anthropic

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/vxcontrol/langchaingo/llms"
)

func truncationLLM(t *testing.T, stopReason string) *LLM {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, r.Body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintf(w, `{"id":"x","type":"message","role":"assistant","model":"claude-3-5-sonnet-latest",
			"content":[{"type":"text","text":"half an ans"}],"stop_reason":%q,
			"usage":{"input_tokens":1,"output_tokens":1}}`, stopReason)
	}))
	t.Cleanup(srv.Close)

	llm, err := New(WithBaseURL(srv.URL), WithToken("unit-test-token"), WithModel("claude-3-5-sonnet-latest"))
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return llm
}

func TestTruncationIsDerivedWithoutRewritingStopReason(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name          string
		stopReason    string
		wantTruncated bool
	}{
		{"out of budget", "max_tokens", true},
		{"finished on its own", "end_turn", false},
		{"hit a stop sequence", "stop_sequence", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			llm := truncationLLM(t, tc.stopReason)

			resp, err := llm.GenerateContent(context.Background(),
				[]llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "hi")})
			if err != nil {
				t.Fatalf("GenerateContent: %v", err)
			}
			if got := resp.Choices[0].Truncated; got != tc.wantTruncated {
				t.Fatalf("Truncated = %v, want %v", got, tc.wantTruncated)
			}
			if got := resp.Choices[0].StopReason; got != tc.stopReason {
				t.Fatalf("StopReason = %q, want the vendor's own %q", got, tc.stopReason)
			}
		})
	}
}

func TestFailOnTruncationKeepsThePartialAnswer(t *testing.T) {
	t.Parallel()

	llm := truncationLLM(t, "max_tokens")
	msgs := []llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "hi")}

	resp, err := llm.GenerateContent(context.Background(), msgs, llms.WithFailOnTruncation())
	if err == nil {
		t.Fatal("want an error once the caller opted in")
	}
	if !llms.IsTruncatedError(err) {
		t.Fatalf("want a truncation error, got %v", err)
	}
	if resp == nil || resp.Choices[0].Content != "half an ans" {
		t.Fatalf("want the partial answer alongside the error, got %+v", resp)
	}

	if _, err := llm.GenerateContent(context.Background(), msgs); err != nil {
		t.Fatalf("without the option the same response must stay a success: %v", err)
	}
}

func legacyCompletionsLLM(t *testing.T, stopReason string) *LLM {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, r.Body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintf(w, `{"completion":"half an ans","model":"claude-2.1","stop_reason":%q}`, stopReason)
	}))
	t.Cleanup(srv.Close)

	llm, err := New(WithBaseURL(srv.URL), WithToken("unit-test-token"),
		WithModel("claude-2.1"), WithLegacyTextCompletionsAPI())
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return llm
}

func TestLegacyCompletionsReportTruncation(t *testing.T) {
	t.Parallel()

	msgs := []llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "hi")}

	t.Run("the stop reason reaches the choice", func(t *testing.T) {
		t.Parallel()
		resp, err := legacyCompletionsLLM(t, "max_tokens").GenerateContent(context.Background(), msgs)
		if err != nil {
			t.Fatalf("GenerateContent: %v", err)
		}
		if got := resp.Choices[0].StopReason; got != "max_tokens" {
			t.Errorf("StopReason = %q, want %q", got, "max_tokens")
		}
		if !resp.Choices[0].Truncated {
			t.Error("a max_tokens stop must mark the choice truncated")
		}
	})

	t.Run("finishing on its own is not truncation", func(t *testing.T) {
		t.Parallel()
		resp, err := legacyCompletionsLLM(t, "stop_sequence").GenerateContent(context.Background(), msgs)
		if err != nil {
			t.Fatalf("GenerateContent: %v", err)
		}
		if resp.Choices[0].Truncated {
			t.Error("a stop_sequence stop must not mark the choice truncated")
		}
	})

	t.Run("failing on truncation keeps the partial answer", func(t *testing.T) {
		t.Parallel()
		resp, err := legacyCompletionsLLM(t, "max_tokens").GenerateContent(context.Background(), msgs,
			llms.WithFailOnTruncation())
		if err == nil {
			t.Fatal("WithFailOnTruncation must fail on a truncated legacy completion")
		}
		if !llms.IsTruncatedError(err) {
			t.Errorf("the error must stay recognizable as truncation, got %v", err)
		}
		if resp == nil || resp.Choices[0].Content != "half an ans" {
			t.Errorf("the partial answer must travel with the error, got %v", resp)
		}
	})
}
