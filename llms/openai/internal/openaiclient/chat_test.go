package openaiclient

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseStreamingChatResponse_FinishReason(t *testing.T) {
	t.Parallel()
	mockBody := `data: {"choices":[{"index":0,"delta":{"role":"assistant","content":"hello"},"finish_reason":"stop"}]}`
	r := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBufferString(mockBody)),
	}

	req := &ChatRequest{
		StreamingFunc: func(ctx context.Context, chunk []byte) error {
			return nil
		},
	}

	resp, err := parseStreamingChatResponse(context.Background(), r, req)

	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "stop", resp.Choices[0].FinishReason)
}
