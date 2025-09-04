package maritacaclient

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStatusError_Error(t *testing.T) {
	tests := []struct {
		name string
		err  StatusError
		want string
	}{
		{
			name: "both status and error message",
			err: StatusError{
				Status:       "400 Bad Request",
				ErrorMessage: "Invalid parameters",
				StatusCode:   400,
			},
			want: "400 Bad Request: Invalid parameters",
		},
		{
			name: "only status",
			err: StatusError{
				Status:     "500 Internal Server Error",
				StatusCode: 500,
			},
			want: "500 Internal Server Error",
		},
		{
			name: "only error message",
			err: StatusError{
				ErrorMessage: "Something went wrong",
				StatusCode:   500,
			},
			want: "Something went wrong",
		},
		{
			name: "empty error",
			err:  StatusError{},
			want: "something went wrong, please see the ollama server logs for details",
		},
		{
			name: "only status code",
			err: StatusError{
				StatusCode: 404,
			},
			want: "something went wrong, please see the ollama server logs for details",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.err.Error())
		})
	}
}

func TestChatRequestStructure(t *testing.T) {
	// Test ChatRequest marshaling
	req := ChatRequest{
		Model: "test-model",
		Messages: []*Message{
			{
				Role:    "user",
				Content: "Hello",
			},
			{
				Role:    "assistant",
				Content: "Hi there!",
			},
		},
		Format: "json",
		Options: Options{
			Token:               "secret-token",
			ChatMode:            true,
			MaxTokens:           100,
			Model:               "sabia-2-medium",
			DoSample:            true,
			Temperature:         0.7,
			TopP:                0.95,
			RepetitionPenalty:   1.0,
			StoppingTokens:      []string{"END", "STOP"},
			Stream:              true,
			NumTokensPerMessage: 4,
		},
	}

	// Verify structure
	assert.Equal(t, "test-model", req.Model)
	assert.Len(t, req.Messages, 2)
	assert.Equal(t, "user", req.Messages[0].Role)
	assert.Equal(t, "Hello", req.Messages[0].Content)
	assert.Equal(t, "json", req.Format)
	assert.Equal(t, true, req.Options.ChatMode)
	assert.Equal(t, 100, req.Options.MaxTokens)
	assert.Equal(t, 0.7, req.Options.Temperature)
}

func TestChatResponseStructure(t *testing.T) {
	resp := ChatResponse{
		Answer: "This is the answer",
		Model:  "test-model",
		Text:   "Generated text",
		Event:  "message",
		Metrics: Metrics{
			Usage: struct {
				CompletionTokens int `json:"completion_tokens"`
				PromptTokens     int `json:"prompt_tokens"`
				TotalTokens      int `json:"total_tokens"`
			}{
				CompletionTokens: 50,
				PromptTokens:     10,
				TotalTokens:      60,
			},
		},
	}

	// Verify structure
	assert.Equal(t, "This is the answer", resp.Answer)
	assert.Equal(t, "test-model", resp.Model)
	assert.Equal(t, "Generated text", resp.Text)
	assert.Equal(t, "message", resp.Event)
	assert.Equal(t, 50, resp.Metrics.Usage.CompletionTokens)
	assert.Equal(t, 10, resp.Metrics.Usage.PromptTokens)
	assert.Equal(t, 60, resp.Metrics.Usage.TotalTokens)
}

func TestMessageStructure(t *testing.T) {
	messages := []Message{
		{
			Role:    "system",
			Content: "You are a helpful assistant",
		},
		{
			Role:    "user",
			Content: "What is the weather?",
		},
		{
			Role:    "assistant",
			Content: "I don't have access to weather data",
		},
	}

	assert.Len(t, messages, 3)
	assert.Equal(t, "system", messages[0].Role)
	assert.Equal(t, "user", messages[1].Role)
	assert.Equal(t, "assistant", messages[2].Role)
}

func TestOptionsDefaults(t *testing.T) {
	// Test with default values
	opts := Options{}

	// These are the zero values, actual defaults would be set by the API
	assert.Equal(t, "", opts.Token)
	assert.Equal(t, false, opts.ChatMode)
	assert.Equal(t, 0, opts.MaxTokens)
	assert.Equal(t, "", opts.Model)
	assert.Equal(t, false, opts.DoSample)
	assert.Equal(t, 0.0, opts.Temperature)
	assert.Equal(t, 0.0, opts.TopP)
	assert.Equal(t, 0.0, opts.RepetitionPenalty)
	assert.Nil(t, opts.StoppingTokens)
	assert.Equal(t, false, opts.Stream)
	assert.Equal(t, 0, opts.NumTokensPerMessage)
}

func TestChatRequestWithStream(t *testing.T) {
	// Test Stream as pointer
	streamTrue := true
	streamFalse := false

	req1 := ChatRequest{
		Model:  "test",
		Stream: &streamTrue,
	}
	assert.NotNil(t, req1.Stream)
	assert.True(t, *req1.Stream)

	req2 := ChatRequest{
		Model:  "test",
		Stream: &streamFalse,
	}
	assert.NotNil(t, req2.Stream)
	assert.False(t, *req2.Stream)

	req3 := ChatRequest{
		Model: "test",
		// Stream is nil
	}
	assert.Nil(t, req3.Stream)
}
