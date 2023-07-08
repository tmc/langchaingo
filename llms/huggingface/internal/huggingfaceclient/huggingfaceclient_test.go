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
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			client, err := New("token", "model")
			require.NoError(t, err)
			// Override the URL to point to our mock server.
			client.url = server.URL

			resp, err := client.RunInference(context.TODO(), tc.req)
			assert.Equal(t, tc.expected, resp)
			if tc.wantErr != "" {
				assert.ErrorContains(t, err, tc.wantErr)
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
			_, _ = w.Write([]byte(fmt.Sprintf(`{"error":["%s"]}`, errMsg)))
		} else {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(fmt.Sprintf(`[{"generated_text":"%s"}]`, goodResponse)))
		}
	}))
}
