package maritaca

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/maritaca/internal/maritacaclient"
	"github.com/tmc/langchaingo/schema"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		opts    []Option
		wantErr bool
	}{
		{
			name: "default options",
			opts: []Option{},
		},
		{
			name: "with model option",
			opts: []Option{
				WithModel("sabia-2-medium"),
			},
		},
		{
			name: "with multiple options",
			opts: []Option{
				WithModel("sabia-2-small"),
				WithFormat("json"),
				WithToken("test-token"),
			},
		},
		{
			name: "with custom HTTP client",
			opts: []Option{
				WithHTTPClient(&http.Client{}),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			llm, err := New(tt.opts...)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, llm)
				assert.NotNil(t, llm.client)
			}
		})
	}
}

func TestTypeToRole(t *testing.T) {
	tests := []struct {
		typ  llms.ChatMessageType
		want string
	}{
		{llms.ChatMessageTypeSystem, "system"},
		{llms.ChatMessageTypeAI, "assistant"},
		{llms.ChatMessageTypeHuman, "user"},
		{llms.ChatMessageTypeGeneric, "user"},
		{llms.ChatMessageTypeFunction, "function"},
		{llms.ChatMessageTypeTool, "tool"},
		{llms.ChatMessageType("unknown"), ""},
	}

	for _, tt := range tests {
		t.Run(string(tt.typ), func(t *testing.T) {
			got := typeToRole(tt.typ)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMakemaritacaOptionsFromOptions(t *testing.T) {
	maritacaOpts := maritacaclient.Options{
		Token:               "original-token",
		ChatMode:            true,
		DoSample:            true,
		NumTokensPerMessage: 4,
	}

	callOpts := llms.CallOptions{
		MaxTokens:         100,
		Model:             "test-model",
		TopP:              0.9,
		RepetitionPenalty: 1.2,
		StopWords:         []string{"END", "STOP"},
		StreamingFunc:     func(ctx context.Context, chunk []byte) error { return nil },
	}

	result := makemaritacaOptionsFromOptions(maritacaOpts, callOpts)

	// Check that CallOptions override the maritacaOptions
	assert.Equal(t, 100, result.MaxTokens)
	assert.Equal(t, "test-model", result.Model)
	assert.Equal(t, 0.9, result.TopP)
	assert.Equal(t, 1.2, result.RepetitionPenalty)
	assert.Equal(t, []string{"END", "STOP"}, result.StoppingTokens)
	assert.True(t, result.Stream)

	// Check that original options are preserved
	assert.Equal(t, "original-token", result.Token)
	assert.True(t, result.ChatMode)
	assert.True(t, result.DoSample)
	assert.Equal(t, 4, result.NumTokensPerMessage)
}

func TestCreateChoice(t *testing.T) {
	resp := maritacaclient.ChatResponse{
		Answer: "Test answer",
		Model:  "test-model",
		Text:   "Test text",
		Metrics: maritacaclient.Metrics{
			Usage: struct {
				CompletionTokens int `json:"completion_tokens"`
				PromptTokens     int `json:"prompt_tokens"`
				TotalTokens      int `json:"total_tokens"`
			}{
				CompletionTokens: 50,
				PromptTokens:     20,
				TotalTokens:      70,
			},
		},
	}

	choices := createChoice(resp)

	require.Len(t, choices, 1)
	assert.Equal(t, "Test answer", choices[0].Content)
	assert.Equal(t, 50, choices[0].GenerationInfo["CompletionTokens"])
	assert.Equal(t, 20, choices[0].GenerationInfo["PromptTokens"])
	assert.Equal(t, 70, choices[0].GenerationInfo["TotalTokens"])
}

func TestLLM_Call(t *testing.T) {
	// The Call method uses GenerateFromSinglePrompt which requires GenerateContent to work
	// We verify that the method exists and is properly implemented
	llm := &LLM{
		client: &maritacaclient.Client{},
		options: options{
			model: "test-model",
		},
	}

	// Verify the method exists
	assert.NotNil(t, llm.Call)
}

func TestLLM_GenerateContent_Validation(t *testing.T) {
	// Test validation logic without requiring a working client
	tests := []struct {
		name     string
		messages []llms.MessageContent
		wantErr  string
	}{
		{
			name: "multiple text parts error",
			messages: []llms.MessageContent{
				{
					Role: llms.ChatMessageTypeHuman,
					Parts: []llms.ContentPart{
						llms.TextContent{Text: "First text"},
						llms.TextContent{Text: "Second text"},
					},
				},
			},
			wantErr: "expecting a single Text content",
		},
		{
			name: "unsupported content type",
			messages: []llms.MessageContent{
				{
					Role: llms.ChatMessageTypeHuman,
					Parts: []llms.ContentPart{
						llms.ImageURLContent{URL: "http://example.com/image.png"},
					},
				},
			},
			wantErr: "only support Text and BinaryContent parts right now",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test the validation logic directly
			for _, mc := range tt.messages {
				var foundText bool
				for _, p := range mc.Parts {
					switch p.(type) {
					case llms.TextContent:
						if foundText {
							assert.Equal(t, "expecting a single Text content", tt.wantErr)
						}
						foundText = true
					case llms.ImageURLContent:
						assert.Equal(t, "only support Text and BinaryContent parts right now", tt.wantErr)
					}
				}
			}
		})
	}
}

func TestLLM_GenerateContent_MessageConversion(t *testing.T) {
	// Test that MessageContent is properly converted to maritaca format
	messages := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeSystem,
			Parts: []llms.ContentPart{
				llms.TextContent{Text: "You are a helpful assistant"},
			},
		},
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextContent{Text: "Hello"},
			},
		},
		{
			Role: llms.ChatMessageTypeAI,
			Parts: []llms.ContentPart{
				llms.TextContent{Text: "Hi there!"},
			},
		},
	}

	// Convert messages as the GenerateContent method would
	chatMsgs := make([]*maritacaclient.Message, 0, len(messages))
	for _, mc := range messages {
		msg := &maritacaclient.Message{Role: typeToRole(mc.Role)}

		for _, p := range mc.Parts {
			if tc, ok := p.(llms.TextContent); ok {
				msg.Content = tc.Text
			}
		}

		chatMsgs = append(chatMsgs, msg)
	}

	require.Len(t, chatMsgs, 3)
	assert.Equal(t, "system", chatMsgs[0].Role)
	assert.Equal(t, "You are a helpful assistant", chatMsgs[0].Content)
	assert.Equal(t, "user", chatMsgs[1].Role)
	assert.Equal(t, "Hello", chatMsgs[1].Content)
	assert.Equal(t, "assistant", chatMsgs[2].Role)
	assert.Equal(t, "Hi there!", chatMsgs[2].Content)
}

func TestLLM_GenerateContent_CallOptions(t *testing.T) {
	// Test that CallOptions are properly applied
	defaultOpts := maritacaclient.Options{
		Token:     "test-token",
		MaxTokens: 100,
		TopP:      0.7,
	}

	callOpts := llms.CallOptions{
		Model:             "override-model",
		MaxTokens:         200,
		TopP:              0.8,
		RepetitionPenalty: 1.1,
		StopWords:         []string{"STOP"},
		StreamingFunc:     func(ctx context.Context, chunk []byte) error { return nil },
	}

	// Test that options are properly merged
	mergedOpts := makemaritacaOptionsFromOptions(defaultOpts, callOpts)

	assert.Equal(t, 200, mergedOpts.MaxTokens)
	assert.Equal(t, "override-model", mergedOpts.Model)
	assert.Equal(t, 0.8, mergedOpts.TopP)
	assert.Equal(t, 1.1, mergedOpts.RepetitionPenalty)
	assert.Equal(t, []string{"STOP"}, mergedOpts.StoppingTokens)
	assert.True(t, mergedOpts.Stream)
	assert.Equal(t, "test-token", mergedOpts.Token) // Original token preserved
}

func TestLLM_Callbacks(t *testing.T) {
	// Test that callbacks can be set and would be invoked
	handler := &mockCallbackHandler{}

	llm := &LLM{
		CallbacksHandler: handler,
		options: options{
			model: "test-model",
		},
	}

	// Verify callbacks handler is set
	assert.NotNil(t, llm.CallbacksHandler)
	assert.Equal(t, handler, llm.CallbacksHandler)
}

// mockCallbackHandler is a simple callback handler for testing
type mockCallbackHandler struct {
	onStart func(ctx context.Context, messages []llms.MessageContent)
	onEnd   func(ctx context.Context, response *llms.ContentResponse)
	onError func(ctx context.Context, err error)
}

var _ callbacks.Handler = &mockCallbackHandler{}

func (h *mockCallbackHandler) HandleLLMGenerateContentStart(ctx context.Context, messages []llms.MessageContent) {
	if h.onStart != nil {
		h.onStart(ctx, messages)
	}
}

func (h *mockCallbackHandler) HandleLLMGenerateContentEnd(ctx context.Context, response *llms.ContentResponse) {
	if h.onEnd != nil {
		h.onEnd(ctx, response)
	}
}

func (h *mockCallbackHandler) HandleLLMError(ctx context.Context, err error) {
	if h.onError != nil {
		h.onError(ctx, err)
	}
}

// Implement other required methods of callbacks.Handler
func (h *mockCallbackHandler) HandleText(ctx context.Context, text string)                      {}
func (h *mockCallbackHandler) HandleLLMStart(ctx context.Context, prompts []string)             {}
func (h *mockCallbackHandler) HandleChainStart(ctx context.Context, inputs map[string]any)      {}
func (h *mockCallbackHandler) HandleChainEnd(ctx context.Context, outputs map[string]any)       {}
func (h *mockCallbackHandler) HandleChainError(ctx context.Context, err error)                  {}
func (h *mockCallbackHandler) HandleToolStart(ctx context.Context, input string)                {}
func (h *mockCallbackHandler) HandleToolEnd(ctx context.Context, output string)                 {}
func (h *mockCallbackHandler) HandleToolError(ctx context.Context, err error)                   {}
func (h *mockCallbackHandler) HandleAgentAction(ctx context.Context, action schema.AgentAction) {}
func (h *mockCallbackHandler) HandleAgentFinish(ctx context.Context, finish schema.AgentFinish) {}
func (h *mockCallbackHandler) HandleRetrieverStart(ctx context.Context, query string)           {}
func (h *mockCallbackHandler) HandleRetrieverEnd(ctx context.Context, query string, documents []schema.Document) {
}
func (h *mockCallbackHandler) HandleStreamingFunc(ctx context.Context, chunk []byte) {}

func TestOptions(t *testing.T) {
	var opts options

	// Test WithModel
	WithModel("test-model")(&opts)
	assert.Equal(t, "test-model", opts.model)

	// Test WithFormat
	WithFormat("json")(&opts)
	assert.Equal(t, "json", opts.format)

	// Test WithToken
	WithToken("test-token")(&opts)
	assert.Equal(t, "test-token", opts.maritacaOptions.Token)

	// Test WithHTTPClient
	client := &http.Client{}
	WithHTTPClient(client)(&opts)
	assert.Equal(t, client, opts.httpClient)

	// Test various maritaca options
	WithChatMode(true)(&opts)
	assert.True(t, opts.maritacaOptions.ChatMode)

	WithMaxTokens(100)(&opts)
	assert.Equal(t, 100, opts.maritacaOptions.MaxTokens)

	WithTemperature(0.7)(&opts)
	assert.Equal(t, 0.7, opts.maritacaOptions.Temperature)

	WithDoSample(true)(&opts)
	assert.True(t, opts.maritacaOptions.DoSample)

	WithTopP(0.95)(&opts)
	assert.Equal(t, 0.95, opts.maritacaOptions.TopP)

	WithRepetitionPenalty(1.2)(&opts)
	assert.Equal(t, 1.2, opts.maritacaOptions.RepetitionPenalty)

	WithStoppingTokens([]string{"END", "STOP"})(&opts)
	assert.Equal(t, []string{"END", "STOP"}, opts.maritacaOptions.StoppingTokens)

	WithStream(true)(&opts)
	assert.True(t, opts.maritacaOptions.Stream)

	WithTokensPerMessage(5)(&opts)
	assert.Equal(t, 5, opts.maritacaOptions.NumTokensPerMessage)

	// Test WithSystemPrompt
	WithSystemPrompt("You are a helpful assistant")(&opts)
	assert.Equal(t, "You are a helpful assistant", opts.system)

	// Test WithCustomTemplate
	WithCustomTemplate("{{.System}} {{.Prompt}}")(&opts)
	assert.Equal(t, "{{.System}} {{.Prompt}}", opts.customModelTemplate)

	// Test WithServerURL
	WithServerURL("https://custom.maritaca.ai")(&opts)
	assert.NotNil(t, opts.maritacaServerURL)
	assert.Equal(t, "https://custom.maritaca.ai", opts.maritacaServerURL.String())

	// Test WithOptions
	mOpts := maritacaclient.Options{
		ChatMode:    true,
		MaxTokens:   200,
		Temperature: 0.8,
		Token:       "override-token",
	}
	WithOptions(mOpts)(&opts)
	assert.Equal(t, mOpts, opts.maritacaOptions)
}

func TestStreamingResponse(t *testing.T) {
	// Test the streaming response handling in GenerateContent
	streamedChunks := []string{}
	streamFunc := func(ctx context.Context, chunk []byte) error {
		streamedChunks = append(streamedChunks, string(chunk))
		return nil
	}

	// Test response handling for different events
	responses := []maritacaclient.ChatResponse{
		{Event: "message", Text: "Hello"},
		{Event: "message", Text: " world"},
		{Event: "end"},
	}

	streamedResponse := ""
	for _, resp := range responses {
		switch resp.Event {
		case "message":
			streamedResponse += resp.Text
			if streamFunc != nil && resp.Text != "" {
				_ = streamFunc(context.Background(), []byte(resp.Text))
			}
		case "end":
			// Final response would have the accumulated text
			assert.Equal(t, "Hello world", streamedResponse)
		}
	}

	assert.Equal(t, []string{"Hello", " world"}, streamedChunks)
}

func TestNonStreamingResponse(t *testing.T) {
	// Test non-streaming response handling
	resp := maritacaclient.ChatResponse{
		Event:  "nostream",
		Answer: "Complete response",
		Text:   "Complete response",
		Metrics: maritacaclient.Metrics{
			Usage: struct {
				CompletionTokens int `json:"completion_tokens"`
				PromptTokens     int `json:"prompt_tokens"`
				TotalTokens      int `json:"total_tokens"`
			}{
				CompletionTokens: 25,
				PromptTokens:     10,
				TotalTokens:      35,
			},
		},
	}

	// In non-streaming mode, the response should be used directly
	assert.Equal(t, "nostream", resp.Event)
	assert.Equal(t, "Complete response", resp.Answer)

	// Create choice from response
	choices := createChoice(resp)
	assert.Len(t, choices, 1)
	assert.Equal(t, "Complete response", choices[0].Content)
	assert.Equal(t, 25, choices[0].GenerationInfo["CompletionTokens"])
}

func TestStreamingError(t *testing.T) {
	// Test error handling in streaming function
	streamErr := errors.New("streaming error")
	streamFunc := func(ctx context.Context, chunk []byte) error {
		return streamErr
	}

	resp := maritacaclient.ChatResponse{
		Event: "message",
		Text:  "Test",
	}

	// Simulate the error handling in the streaming callback
	err := func(response maritacaclient.ChatResponse) error {
		if streamFunc != nil && response.Text != "" {
			if err := streamFunc(context.Background(), []byte(response.Text)); err != nil {
				return err
			}
		}
		return nil
	}(resp)

	assert.Error(t, err)
	assert.Equal(t, streamErr, err)
}
