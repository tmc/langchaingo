package perplexityclient

import (
	"bytes"
	"context"
	"encoding/json"
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

func TestChatMessage_MarshalUnmarshal(t *testing.T) {
	t.Parallel()
	msg := ChatMessage{
		Role:    "assistant",
		Content: "hello",
		FunctionCall: &FunctionCall{
			Name:      "test",
			Arguments: "func",
		},
	}
	text, err := json.Marshal(msg)
	require.NoError(t, err)
	require.Equal(t, `{"role":"assistant","content":"hello","function_call":{"name":"test","arguments":"func"}}`, string(text)) // nolint: lll

	var msg2 ChatMessage
	err = json.Unmarshal(text, &msg2)
	require.NoError(t, err)
	require.Equal(t, msg, msg2)
}
