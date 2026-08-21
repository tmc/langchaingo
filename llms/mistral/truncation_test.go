package mistral

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/llms/streaming"
)

func truncationModel(t *testing.T, finishReason string) *Model {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, r.Body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintf(w, `{"id":"x","object":"chat.completion","created":1,"model":"mistral-small-latest","choices":[{"index":0,"message":{"role":"assistant","content":"half an ans"},"finish_reason":%q}],"usage":{"prompt_tokens":1,"completion_tokens":1,"total_tokens":2}}`, finishReason)
	}))
	t.Cleanup(srv.Close)

	model, err := New(WithAPIKey("unit-test-token"), WithEndpoint(srv.URL), WithModel("mistral-small-latest"))
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return model
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
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			model := truncationModel(t, tc.finishReason)

			resp, err := model.GenerateContent(context.Background(),
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

	model := truncationModel(t, "length")
	msgs := []llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "hi")}

	resp, err := model.GenerateContent(context.Background(), msgs, llms.WithFailOnTruncation())
	if err == nil {
		t.Fatal("want an error once the caller opted in")
	}
	if !llms.IsTruncatedError(err) {
		t.Fatalf("want a truncation error, got %v", err)
	}
	if resp == nil || resp.Choices[0].Content != "half an ans" {
		t.Fatalf("want the partial answer alongside the error, got %+v", resp)
	}

	if _, err := model.GenerateContent(context.Background(), msgs); err != nil {
		t.Fatalf("without the option the same response must stay a success: %v", err)
	}
}

func TestTruncationOnTheStreamingPath(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, r.Body)
		w.Header().Set("Content-Type", "text/event-stream")
		flusher, _ := w.(http.Flusher)
		for _, chunk := range []string{
			`{"id":"x","object":"chat.completion.chunk","created":1,"model":"mistral-small-latest","choices":[{"index":0,"delta":{"role":"assistant","content":"half an ans"},"finish_reason":""}]}`,
			`{"id":"x","object":"chat.completion.chunk","created":1,"model":"mistral-small-latest","choices":[{"index":0,"delta":{"content":""},"finish_reason":"length"}]}`,
		} {
			_, _ = fmt.Fprintf(w, "data: %s\n\n", chunk)
			if flusher != nil {
				flusher.Flush()
			}
		}
		_, _ = io.WriteString(w, "data: [DONE]\n\n")
	}))
	t.Cleanup(srv.Close)

	model, err := New(WithAPIKey("unit-test-token"), WithEndpoint(srv.URL), WithModel("mistral-small-latest"))
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	msgs := []llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "hi")}
	sink := func(context.Context, streaming.Chunk) error { return nil }

	resp, err := model.GenerateContent(context.Background(), msgs, llms.WithStreamingFunc(sink))
	if err != nil {
		t.Fatalf("GenerateContent: %v", err)
	}
	if !resp.Choices[0].Truncated {
		t.Fatalf("Truncated = false on a streamed answer that hit the limit, StopReason = %q", resp.Choices[0].StopReason)
	}
	if got := resp.Choices[0].StopReason; got != "length" {
		t.Fatalf("StopReason = %q, want the vendor's own %q", got, "length")
	}
}
