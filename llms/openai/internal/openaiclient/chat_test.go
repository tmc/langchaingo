package openaiclient

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
	ctx := context.Background()
	t.Parallel()
	mockBody := `data: {"choices":[{"index":0,"delta":{"role":"assistant","content":"hello"},"finish_reason":"stop"}]}`
	r := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBufferString(mockBody)),
	}

	req := &ChatRequest{
		StreamingFunc: func(_ context.Context, _ []byte) error {
			return nil
		},
	}

	resp, err := parseStreamingChatResponse(ctx, r, req)

	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, FinishReason("stop"), resp.Choices[0].FinishReason)
}

func TestParseStreamingChatResponse_ReasoningContent(t *testing.T) {
	ctx := context.Background()
	t.Parallel()
	mockBody := `data: {"choices":[{"index":0,"delta":{"role":"assistant","content":"final answer","reasoning_content":"step-by-step reasoning"},"finish_reason":"stop"}]}`
	r := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBufferString(mockBody)),
	}

	req := &ChatRequest{
		StreamingFunc: func(_ context.Context, _ []byte) error {
			return nil
		},
	}

	resp, err := parseStreamingChatResponse(ctx, r, req)

	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "final answer", resp.Choices[0].Message.Content)
	assert.Equal(t, "step-by-step reasoning", resp.Choices[0].Message.ReasoningContent)
	assert.Equal(t, FinishReason("stop"), resp.Choices[0].FinishReason)
}

func TestParseStreamingChatResponse_ReasoningFunc(t *testing.T) {
	ctx := context.Background()
	t.Parallel()
	mockBody := `
data: {"id":"fa7e4fc5-a05d-4e7b-9a66-a2dd89e91a4e","object":"chat.completion.chunk","created":1738492867,"model":"deepseek-reasoner","system_fingerprint":"fp_7e73fd9a08","choices":[{"index":0,"delta":{"content":null,"reasoning_content":"Okay"},"logprobs":null,"finish_reason":null}]}
`
	r := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBufferString(mockBody)),
	}

	req := &ChatRequest{
		StreamingReasoningFunc: func(_ context.Context, reasoningChunk, chunk []byte) error {
			t.Logf("reasoningChunk: %s", string(reasoningChunk))
			t.Logf("chunk: %s", string(chunk))
			return nil
		},
	}

	resp, err := parseStreamingChatResponse(ctx, r, req)

	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "", resp.Choices[0].Message.Content)
	assert.Equal(t, "Okay", resp.Choices[0].Message.ReasoningContent)
	assert.Equal(t, FinishReason(""), resp.Choices[0].FinishReason)
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

func TestChatMessage_MarshalUnmarshal_WithReasoning(t *testing.T) {
	t.Parallel()
	msg := ChatMessage{
		Role:             "assistant",
		Content:          "final answer",
		ReasoningContent: "step-by-step reasoning",
	}
	text, err := json.Marshal(msg)
	require.NoError(t, err)
	require.Equal(t, `{"role":"assistant","content":"final answer","reasoning_content":"step-by-step reasoning"}`, string(text))

	var msg2 ChatMessage
	err = json.Unmarshal(text, &msg2)
	require.NoError(t, err)
	require.Equal(t, msg, msg2)
}

func TestParseStreamingChatResponse_StreamingToolCalls(t *testing.T) {
	ctx := context.Background()
	t.Parallel()

	message := `
data: {"choices":[{"index":0,"delta":{"role":"assistant","tool_calls":[{"function":{"name":"test","arguments":"func"}}]}}]}
data: {"choices":[{"index":0,"delta":{"role":"assistant","tool_calls":[{"function":{"arguments":"tion"}}]}}]}
data: {"choices":[{"index":0,"delta":{"role":"assistant","tool_calls":[{"function":{"arguments":" call"}}]}}]}
	`

	r := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBufferString(message)),
	}

	req := &ChatRequest{
		StreamingFunc: func(_ context.Context, _ []byte) error {
			return nil
		},
	}

	resp, err := parseStreamingChatResponse(ctx, r, req)

	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp.Choices[0].Message.ToolCalls, 1)
	assert.Equal(t, resp.Choices[0].Message.ToolCalls[0].Function.Name, "test")
	assert.Equal(t, resp.Choices[0].Message.ToolCalls[0].Function.Arguments, "function call")
}

func TestParseStreamingChatResponse_StreamingMultipleToolCalls(t *testing.T) {
	ctx := context.Background()
	t.Parallel()

	message := `
data: {"choices":[{"index":0,"delta":{"role":"assistant","tool_calls":[{"function":{"name":"test","arguments":"func"}}]}}]}
data: {"choices":[{"index":0,"delta":{"role":"assistant","tool_calls":[{"function":{"arguments":"tion"}}]}}]}
data: {"choices":[{"index":0,"delta":{"role":"assistant","tool_calls":[{"function":{"arguments":" call"}}]}}]}
data: {"choices":[{"index":0,"delta":{"role":"assistant","tool_calls":[{"function":{"name":"test2","arguments":"other"}}]}}]}
data: {"choices":[{"index":0,"delta":{"role":"assistant","tool_calls":[{"function":{"arguments":" function"}}]}}]}
data: {"choices":[{"index":0,"delta":{"role":"assistant","tool_calls":[{"function":{"arguments":" call"}}]}}]}
	`

	r := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBufferString(message)),
	}

	req := &ChatRequest{
		StreamingFunc: func(_ context.Context, _ []byte) error {
			return nil
		},
	}

	resp, err := parseStreamingChatResponse(ctx, r, req)

	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp.Choices[0].Message.ToolCalls, 2)

	assert.Equal(t, resp.Choices[0].Message.ToolCalls[0].Function.Name, "test")
	assert.Equal(t, resp.Choices[0].Message.ToolCalls[0].Function.Arguments, "function call")

	assert.Equal(t, resp.Choices[0].Message.ToolCalls[1].Function.Name, "test2")
	assert.Equal(t, resp.Choices[0].Message.ToolCalls[1].Function.Arguments, "other function call")
}

func TestParseStreamingChatResponse_StreamingMultipleToolCalls_BackwardsCompatibility(t *testing.T) {
	ctx := context.Background()
	t.Parallel()

	message := `
data: {"choices":[{"index":0,"delta":{"role":"assistant","tool_calls":[{"type":"function", "function":{"name":"test","arguments":"func"}}]}}]}
data: {"choices":[{"index":0,"delta":{"role":"assistant","tool_calls":[{"function":{"arguments":"tion"}}]}}]}
data: {"choices":[{"index":0,"delta":{"role":"assistant","tool_calls":[{"function":{"arguments":" call"}}]}}]}
data: {"choices":[{"index":0,"delta":{"role":"assistant","tool_calls":[{"type":"function", "function":{"name":"test2","arguments":"other"}}]}}]}
data: {"choices":[{"index":0,"delta":{"role":"assistant","tool_calls":[{"function":{"arguments":" function"}}]}}]}
data: {"choices":[{"index":0,"delta":{"role":"assistant","tool_calls":[{"function":{"arguments":" call"}}]}}]}
	`

	r := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBufferString(message)),
	}

	req := &ChatRequest{
		StreamingFunc: func(_ context.Context, _ []byte) error {
			return nil
		},
	}

	resp, err := parseStreamingChatResponse(ctx, r, req)

	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp.Choices[0].Message.ToolCalls, 2)

	assert.Equal(t, resp.Choices[0].Message.ToolCalls[0].Function.Name, "test")
	assert.Equal(t, resp.Choices[0].Message.ToolCalls[0].Function.Arguments, "function call")

	assert.Equal(t, resp.Choices[0].Message.ToolCalls[1].Function.Name, "test2")
	assert.Equal(t, resp.Choices[0].Message.ToolCalls[1].Function.Arguments, "other function call")
}
