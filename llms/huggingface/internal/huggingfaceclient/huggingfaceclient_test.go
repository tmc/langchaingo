package huggingfaceclient

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/llms"
)

const (
	goodResponse = "I hug, you hug, we all hug!"
	errMsg       = "Error in `parameters.top_k`: ensure this value is greater than or equal to 1"
)

func TestRunInference(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		req      *InferenceRequest
		expected *InferenceResponse
		wantErr  string
	}{
		{"ok", &InferenceRequest{}, &InferenceResponse{Text: goodResponse}, ""},
		{"stream ok", &InferenceRequest{Stream: true}, &InferenceResponse{Text: goodResponse}, ""},
		{"not ok", &InferenceRequest{TopK: -1}, nil, errMsg},
	}

	server := mockServer(t)
	t.Cleanup(server.Close)

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			client, err := New("token", "model", server.URL)
			require.NoError(t, err)
			var streamResponse string
			resp, err := client.RunInference(context.TODO(), tc.req, &llms.CallOptions{
				StreamingFunc: func(_ context.Context, chunk []byte) error {
					streamResponse += string(chunk)
					return nil
				},
			})
			assert.Equal(t, tc.expected, resp)
			if tc.req.Stream {
				assert.Equal(t, tc.expected.Text, streamResponse)
			}
			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)
			}
		})
	}
}

func mockServer(t *testing.T) *httptest.Server {
	t.Helper()

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, err := io.ReadAll(r.Body)
		require.NoError(t, err)

		var infReq inferencePayload
		err = json.Unmarshal(b, &infReq)
		require.NoError(t, err)

		switch {
		case infReq.Parameters.TopK == -1:
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(fmt.Sprintf(`{"error":["%s"]}`, errMsg)))
		case infReq.Stream:
			// set Content-Type text/event-stream
			w.Header().Set("Content-Type", "text/event-stream")
			w.Header().Set("Cache-Control", "no-cache")
			w.Header().Set("Connection", "keep-alive")

			flusher, ok := w.(http.Flusher)
			if !ok {
				http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
				return
			}
			fmt.Fprintf(w, `data: {"index":6,"token":{"id":27121,"text":"I hug, you hug, ","logprob":-0.023592351,"special":false},"generated_text":null,"details":null}`+"\n\n")
			flusher.Flush()
			fmt.Fprintf(w, `data: {"index":6,"token":{"id":27121,"text":"we all hug!","logprob":-0.023592351,"special":false},"generated_text":"%s","details":null}`+"\n\n", goodResponse)
			flusher.Flush()
		default:
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(fmt.Sprintf(`[{"generated_text":"%s"}]`, goodResponse)))
		}
	}))
}
