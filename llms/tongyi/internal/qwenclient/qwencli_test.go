package qwenclient

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func newQwenClient(t *testing.T, model string) *QwenClient {
	t.Helper()
	cli := NewQwenClient(model, NewHTTPClient())
	if cli.token == "" {
		t.Skip("token is empty")
	}
	return cli
}

func newMockClient(t *testing.T, model string, ctrl *gomock.Controller, f mockFn) *QwenClient {
	t.Helper()

	mockHTTPCli := NewMockIHttpClient(ctrl)
	f(mockHTTPCli)

	qwenCli := NewQwenClient(model, mockHTTPCli)
	return qwenCli
}

type mockFn func(mockHTTPCli *MockIHttpClient)

func TestStreamingChunk(t *testing.T) {
	t.Parallel()
	ctx := context.TODO()

	cli := newQwenClient(t, "qwen-turbo")

	output := ""

	input := Input{
		Messages: []Message{
			{Role: "user", Content: "Hello!"},
		},
	}

	req := &QwenRequest{
		Input: input,
		StreamingFunc: func(ctx context.Context, chunk []byte) error {
			output += string(chunk)
			return nil
		},
	}
	resp, err := cli.CreateCompletion(ctx, req)

	require.NoError(t, err)
	assert.Regexp(t, "hello|hi|how|assist", resp.Output.Choices[0].Message.Content)
	assert.Regexp(t, "hello|hi|how|assist", output)
}

func TestMockStreamingChunk(t *testing.T) {
	t.Parallel()
	ctx := context.TODO()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	cli := newMockClient(t, "qwen-turbo", ctrl, _mockAsyncFunc)

	output := ""

	input := Input{
		Messages: []Message{
			{Role: "user", Content: "Hello!"},
		},
	}

	req := &QwenRequest{
		Input: input,
		StreamingFunc: func(ctx context.Context, chunk []byte) error {
			output += string(chunk)
			return nil
		},
	}
	resp, err := cli.CreateCompletion(ctx, req)

	require.NoError(t, err)

	assert.Equal(t, "Hello! How can I assist you today?", resp.Output.Choices[0].Message.Content)
	assert.Equal(t, "Hello! How can I assist you today?", output)
}

func TestMockBasic(t *testing.T) {
	t.Parallel()
	ctx := context.TODO()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	cli := newMockClient(t, "qwen-turbo", ctrl, _mockSyncFunc)
	input := Input{
		Messages: []Message{
			{Role: "user", Content: "Hello!"},
		},
	}

	req := &QwenRequest{
		Input: input,
	}

	resp, err := cli.CreateCompletion(ctx, req)

	require.NoError(t, err)

	assert.Equal(t, "Hello! This is a mock message.", resp.Output.Choices[0].Message.Content)
	assert.Equal(t, "mock-ac55-9fd3-8326-8415cbdf5683", resp.RequestID)
	assert.Equal(t, 15, resp.Usage.TotalTokens)
}

func _mockAsyncFunc(mockHTTPCli *MockIHttpClient) {
	MockStreamData := []string{
		`id:1`,
		`event:result`,
		`:HTTP_STATUS/200`,
		`data:{
			"output": {
				"choices": [{
					"message": {
						"content": "Hello! How",
						"role": "assistant"
					},
					"finish_reason": "null"
				}]
			},
			"usage": {
				"total_tokens": 9,
				"input_tokens": 6,
				"output_tokens": 3
			},
			"request_id": "95bea986-ac55-9fd3-8326-8415cbdf5683"
		}`,
		`    `,
		`id:2`,
		`event:result`,
		`:HTTP_STATUS/200`,
		`data:{
			"output": {
				"choices": [{
					"message": {
						"content": " can I assist you today?",
						"role": "assistant"
					},
					"finish_reason": "null"
				}]
			},
			"usage": {
				"total_tokens": 15,
				"input_tokens": 6,
				"output_tokens": 9
			},
			"request_id": "95bea986-ac55-9fd3-8326-8415cbdf5683"
		}`,
		`    `,
		`id:3`,
		`event:result`,
		`:HTTP_STATUS/200`,
		`data:{
			"output": {
				"choices": [{
					"message": {
						"content": "",
						"role": "assistant"
					},
					"finish_reason": "stop"
				}]
			},
			"usage": {
				"total_tokens": 15,
				"input_tokens": 6,
				"output_tokens": 9
			},
			"request_id": "95bea986-ac55-9fd3-8326-8415cbdf5683"
		}`,
	}

	ctx := context.TODO()

	_rawStreamOutChannel := make(chan string, 500)

	mockHTTPCli.EXPECT().
		PostSSE(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(_rawStreamOutChannel, nil)
	go func() {
		for _, line := range MockStreamData {
			_rawStreamOutChannel <- line
		}
		close(_rawStreamOutChannel)
	}()
}

func _mockSyncFunc(mockHTTPCli *MockIHttpClient) {
	ctx := context.TODO()

	mockResp := QwenOutputMessage{
		Output: QwenOutput{
			Choices: []struct {
				Message      Message `json:"message"`
				FinishReason string  `json:"finish_reason"`
			}{
				{
					Message: Message{
						Content: "Hello! This is a mock message.",
						Role:    "assistant",
					},
					FinishReason: "stop",
				},
			},
		},
		Usage: struct {
			TotalTokens  int `json:"total_tokens"`
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
		}{
			TotalTokens:  15,
			InputTokens:  6,
			OutputTokens: 9,
		},
		RequestID: "mock-ac55-9fd3-8326-8415cbdf5683",
	}
	mockHTTPCli.EXPECT().
		Post(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		SetArg(3, mockResp).
		Return(nil)
}
