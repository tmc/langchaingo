package openai

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/vxcontrol/langchaingo/llms"
)

func truncationServer(t *testing.T, finishReason string) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, r.Body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintf(w, `{"id":"x","object":"chat.completion","created":1,"model":"gpt-4o",
			"choices":[{"index":0,"message":{"role":"assistant","content":"half an ans"},
			"finish_reason":%q}],
			"usage":{"prompt_tokens":1,"completion_tokens":1,"total_tokens":2}}`, finishReason)
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestTruncationIsDerivedWithoutRewritingStopReason(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name          string
		finishReason  string
		wantTruncated bool
	}{
		{"out of budget", "length", true},
		{"finished on its own", "stop", false},
		{"asked for a tool", "tool_calls", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			srv := truncationServer(t, tc.finishReason)
			llm := newUnitLLM(t, WithBaseURL(srv.URL), WithModel("gpt-4o"))

			resp, err := llm.GenerateContent(context.Background(),
				[]llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "hi")})
			if err != nil {
				t.Fatalf("GenerateContent: %v", err)
			}
			if got := resp.Choices[0].Truncated; got != tc.wantTruncated {
				t.Fatalf("Truncated = %v, want %v", got, tc.wantTruncated)
			}
			if got := resp.Choices[0].StopReason; got != tc.finishReason {
				t.Fatalf("StopReason = %q, want the vendor's own %q", got, tc.finishReason)
			}
		})
	}
}

func TestFailOnTruncationKeepsThePartialAnswer(t *testing.T) {
	t.Parallel()

	srv := truncationServer(t, "length")
	llm := newUnitLLM(t, WithBaseURL(srv.URL), WithModel("gpt-4o"))
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
