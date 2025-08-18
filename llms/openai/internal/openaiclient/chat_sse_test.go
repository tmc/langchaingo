package openaiclient

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"
)

func TestParseStreamingChatResponse_SSEComments(t *testing.T) {
	ctx := context.Background()
	t.Parallel()

	// Test the key SSE comment patterns
	testCases := []struct {
		name            string
		body            string
		expectedContent string
	}{
		{
			name: "openrouter_comments",
			body: `data: {"id":"1","object":"chat.completion.chunk","created":1234567890,"model":"test","choices":[{"index":0,"delta":{"content":"Hello"},"finish_reason":null}]}
: OPENROUTER PROCESSING
: OPENROUTER PROCESSING
data: {"id":"1","object":"chat.completion.chunk","created":1234567890,"model":"test","choices":[{"index":0,"delta":{"content":" World"},"finish_reason":null}]}
data: {"id":"1","object":"chat.completion.chunk","created":1234567890,"model":"test","choices":[{"index":0,"delta":{},"finish_reason":"stop"}]}
data: [DONE]`,
			expectedContent: "Hello World",
		},
		{
			name: "comments_without_space",
			body: `data: {"id":"1","object":"chat.completion.chunk","created":1234567890,"model":"test","choices":[{"index":0,"delta":{"content":"Test"},"finish_reason":null}]}
:comment-without-space
data: {"id":"1","object":"chat.completion.chunk","created":1234567890,"model":"test","choices":[{"index":0,"delta":{},"finish_reason":"stop"}]}
data: [DONE]`,
			expectedContent: "Test",
		},
		{
			name: "other_sse_fields",
			body: `event: message
id: 12345
data: {"id":"1","object":"chat.completion.chunk","created":1234567890,"model":"test","choices":[{"index":0,"delta":{"content":"Data"},"finish_reason":null}]}
retry: 1000
data: {"id":"1","object":"chat.completion.chunk","created":1234567890,"model":"test","choices":[{"index":0,"delta":{},"finish_reason":"stop"}]}
data: [DONE]`,
			expectedContent: "Data",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			
			r := &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBufferString(tc.body)),
			}

			req := &ChatRequest{
				StreamingFunc: func(_ context.Context, _ []byte) error {
					return nil
				},
			}

			resp, err := parseStreamingChatResponse(ctx, r, req)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if resp == nil {
				t.Fatal("response should not be nil")
			}
			if len(resp.Choices) == 0 {
				t.Fatal("expected at least one choice")
			}
			if got := resp.Choices[0].Message.Content; got != tc.expectedContent {
				t.Errorf("content mismatch: got %q, want %q", got, tc.expectedContent)
			}
		})
	}
}