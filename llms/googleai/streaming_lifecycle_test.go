package googleai

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/tmc/langchaingo/llms"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestStreamingCallbackError(t *testing.T) {
	wantErr := errors.New("stop streaming")
	client := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": {"application/json"}},
			Body: io.NopCloser(strings.NewReader(`[
				{"candidates":[{"content":{"parts":[{"text":"one"}],"role":"model"}}]},
				{"candidates":[{"content":{"parts":[{"text":"two"}],"role":"model"},"finishReason":1}]}
			]`)),
		}, nil
	})}

	model, err := New(context.Background(), WithRest(), WithAPIKey("test-key"), WithHTTPClient(client))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = model.Close() })

	response, err := model.GenerateContent(context.Background(), []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeHuman, "hello"),
	}, llms.WithStreamingFunc(func(context.Context, []byte) error {
		return wantErr
	}))
	if !errors.Is(err, wantErr) {
		t.Fatalf("GenerateContent error = %v, want %v", err, wantErr)
	}
	if response != nil {
		t.Fatalf("GenerateContent response = %#v, want nil", response)
	}
}
