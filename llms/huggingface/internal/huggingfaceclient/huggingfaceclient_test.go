package huggingfaceclient

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/internal/httprr"
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
		{"not ok", &InferenceRequest{TopK: -1}, nil, errMsg},
	}

	server := mockServer(t)
	t.Cleanup(server.Close)

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			client, err := New("token", "model", server.URL)
			require.NoError(t, err)

			resp, err := client.RunInference(context.TODO(), tc.req)
			assert.Equal(t, tc.expected, resp)
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

		if infReq.Parameters.TopK == -1 {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = fmt.Fprintf(w, `{"error":["%s"]}`, errMsg)
		} else {
			w.WriteHeader(http.StatusOK)
			_, _ = fmt.Fprintf(w, `[{"generated_text":"%s"}]`, goodResponse)
		}
	}))
}

func TestRunInferenceWithHTTPrr(t *testing.T) {
	t.Parallel()

	if os.Getenv("HUGGINGFACE_API_TOKEN") == "" {
		t.Skip("HUGGINGFACE_API_TOKEN not set")
	}

	// Enable recording mode with TEST_RECORD=true
	mode := httprr.ModeReplay
	if os.Getenv("TEST_RECORD") == "true" {
		mode = httprr.ModeRecord
	}

	httpClient := httprr.Client("testdata/huggingface_inference.json", mode)
	
	client, err := New(os.Getenv("HUGGINGFACE_API_TOKEN"), "microsoft/DialoGPT-medium", "")
	require.NoError(t, err)
	
	// Set the HTTPrr client
	client.SetHTTPClient(httpClient)

	req := &InferenceRequest{
		Model:       "microsoft/DialoGPT-medium",
		Prompt:      "What is machine learning?",
		Task:        InferenceTaskTextGeneration,
		Temperature: 0.7,
		MaxLength:   100,
	}

	resp, err := client.RunInference(context.Background(), req)
	require.NoError(t, err)
	assert.NotEmpty(t, resp.Text)
	assert.Contains(t, resp.Text, "machine learning")
}
