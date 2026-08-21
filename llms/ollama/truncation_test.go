package ollama

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/vxcontrol/langchaingo/llms"
)

func truncationLLM(t *testing.T, doneReason string) *LLM {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, r.Body)
		w.Header().Set("Content-Type", "application/x-ndjson")
		_, _ = fmt.Fprintf(w, `{"model":"llama3","created_at":"2026-08-21T09:00:00Z","message":{"role":"assistant","content":"half an ans"},"done":true,"done_reason":%q,"eval_count":1,"prompt_eval_count":1}`+"\n", doneReason)
	}))
	t.Cleanup(srv.Close)

	llm, err := New(WithServerURL(srv.URL), WithModel("llama3"))
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return llm
}

func TestTruncationIsDerivedWithoutRewritingStopReason(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name          string
		doneReason    string
		wantTruncated bool
	}{
		{"out of budget", "length", true},
		{"finished on its own", "stop", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			llm := truncationLLM(t, tc.doneReason)

			resp, err := llm.GenerateContent(context.Background(),
				[]llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "hi")})
			if err != nil {
				t.Fatalf("GenerateContent: %v", err)
			}
			if got := resp.Choices[0].Truncated; got != tc.wantTruncated {
				t.Fatalf("Truncated = %v, want %v", got, tc.wantTruncated)
			}
			if got := resp.Choices[0].StopReason; got != tc.doneReason {
				t.Fatalf("StopReason = %q, want the vendor's own %q", got, tc.doneReason)
			}
		})
	}
}

func TestFailOnTruncationKeepsThePartialAnswer(t *testing.T) {
	t.Parallel()

	llm := truncationLLM(t, "length")
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
