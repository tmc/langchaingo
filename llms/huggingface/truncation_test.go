package huggingface

import (
	"context"
	"errors"
	"github.com/vxcontrol/langchaingo/callbacks"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/vxcontrol/langchaingo/llms"
)

func serveJSON(t *testing.T, body string) (*httptest.Server, *string) {
	t.Helper()
	var request string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		request = string(b)
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, body)
	}))
	t.Cleanup(srv.Close)
	return srv, &request
}

type rewriteHost struct{ target *url.URL }

func (h rewriteHost) RoundTrip(r *http.Request) (*http.Response, error) {
	r.URL.Scheme, r.URL.Host = h.target.Scheme, h.target.Host
	return http.DefaultTransport.RoundTrip(r)
}

func TestTruncationOnTheInferenceDoor(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name      string
		body      string
		reason    string
		truncated bool
	}{
		{"hit the limit", `[{"generated_text":"partial","details":{"finish_reason":"length"}}]`, "length", true},
		{"finished", `[{"generated_text":"done","details":{"finish_reason":"eos_token"}}]`, "eos_token", false},
		{"no details", `[{"generated_text":"done"}]`, "", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			srv, request := serveJSON(t, tc.body)
			llm, err := New(WithToken("t"), WithURL(srv.URL), WithModel("m"))
			if err != nil {
				t.Fatalf("New() error: %v", err)
			}
			resp, err := llm.GenerateContent(context.Background(),
				[]llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "hi")})
			if err != nil {
				t.Fatalf("GenerateContent() error: %v", err)
			}
			if got := resp.Choices[0].StopReason; got != tc.reason {
				t.Errorf("StopReason = %q, want %q", got, tc.reason)
			}
			if got := resp.Choices[0].Truncated; got != tc.truncated {
				t.Errorf("Truncated = %v, want %v", got, tc.truncated)
			}
			if !strings.Contains(*request, `"details":true`) {
				t.Errorf("the door reports a finish reason only when details are requested, got body: %s", *request)
			}
		})
	}
}

func TestTruncationOnTheRouterDoor(t *testing.T) {
	t.Parallel()

	const body = `{"choices":[{"index":0,"message":{"content":"partial"},"finish_reason":"length"}]}`
	srv, _ := serveJSON(t, body)
	// The router path pins its own URL, so the test server is reached by transport.
	target, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatalf("parse test server url: %v", err)
	}
	redirect := &http.Client{Transport: rewriteHost{target: target}}
	llm, err := New(WithToken("t"), WithModel("m"),
		WithInferenceProvider("together"), WithHTTPClient(redirect))
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	resp, err := llm.GenerateContent(context.Background(),
		[]llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "hi")})
	if err != nil {
		t.Fatalf("GenerateContent() error: %v", err)
	}
	if resp.Choices[0].StopReason != "length" || !resp.Choices[0].Truncated {
		t.Fatalf("router door must report truncation, got %q / %v",
			resp.Choices[0].StopReason, resp.Choices[0].Truncated)
	}
}

func TestFailOnTruncationOnTheInferenceDoor(t *testing.T) {
	t.Parallel()

	srv, _ := serveJSON(t, `[{"generated_text":"partial","details":{"finish_reason":"length"}}]`)
	llm, err := New(WithToken("t"), WithURL(srv.URL), WithModel("m"))
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	handler := &recordingCallbacks{}
	llm.CallbacksHandler = handler

	resp, err := llm.GenerateContent(context.Background(),
		[]llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "hi")},
		llms.WithFailOnTruncation())
	if err == nil {
		t.Fatal("a caller that opted in must get a typed error on a truncated answer")
	}
	var apiErr *llms.Error
	if !errors.As(err, &apiErr) || apiErr.Code != llms.ErrCodeTruncated {
		t.Fatalf("expected ErrCodeTruncated, got: %v", err)
	}
	if resp == nil || len(resp.Choices) == 0 {
		t.Fatal("the partial answer must travel with the error, as it does on every other door")
	}
	if got := resp.Choices[0].Content; got != "partial" {
		t.Fatalf("partial answer = %q, want %q", got, "partial")
	}
	if handler.errors != 1 {
		t.Fatalf("HandleLLMError calls = %d, want 1", handler.errors)
	}
}

type recordingCallbacks struct {
	callbacks.SimpleHandler
	errors int
}

func (h *recordingCallbacks) HandleLLMError(context.Context, error) { h.errors++ }
