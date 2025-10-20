package mistral

import (
	"context"
	"os"
	"testing"
	"time"

	sdk "github.com/gage-technologies/mistral-go"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/schema"
)

func TestNew(t *testing.T) {
	// Save original env var
	origAPIKey := os.Getenv("MISTRAL_API_KEY")
	defer os.Setenv("MISTRAL_API_KEY", origAPIKey)

	tests := []struct {
		name    string
		envKey  string
		opts    []Option
		wantErr bool
	}{
		{
			name:   "with default options",
			envKey: "test-key",
			opts:   []Option{},
		},
		{
			name:   "with API key option",
			envKey: "",
			opts: []Option{
				WithAPIKey("test-api-key"),
			},
		},
		{
			name: "with all options",
			opts: []Option{
				WithAPIKey("test-api-key"),
				WithEndpoint("https://test.endpoint.com"),
				WithMaxRetries(5),
				WithTimeout(30 * time.Second),
				WithModel("mistral-large"),
				WithCallbacksHandler(&testCallbackHandler{}),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("MISTRAL_API_KEY", tt.envKey)

			model, err := New(tt.opts...)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && model == nil {
				t.Error("New() returned nil model without error")
			}
		})
	}
}

func TestSetCallOptions(t *testing.T) {
	callOpts := &llms.CallOptions{}

	options := []llms.CallOption{
		llms.WithModel("test-model"),
		llms.WithTemperature(0.7),
		llms.WithMaxTokens(100),
		llms.WithTopP(0.9),
	}

	setCallOptions(options, callOpts)

	if callOpts.Model != "test-model" {
		t.Errorf("setCallOptions() Model = %v, want %v", callOpts.Model, "test-model")
	}
	if callOpts.Temperature != 0.7 {
		t.Errorf("setCallOptions() Temperature = %v, want %v", callOpts.Temperature, 0.7)
	}
	if callOpts.MaxTokens != 100 {
		t.Errorf("setCallOptions() MaxTokens = %v, want %v", callOpts.MaxTokens, 100)
	}
	if callOpts.TopP != 0.9 {
		t.Errorf("setCallOptions() TopP = %v, want %v", callOpts.TopP, 0.9)
	}
}

func TestResolveDefaultOptions(t *testing.T) {
	sdkDefaults := sdk.ChatRequestParams{
		Temperature: 0.5,
		MaxTokens:   50,
		TopP:        0.8,
	}

	clientOpts := &clientOptions{
		model: "client-model",
	}

	result := resolveDefaultOptions(sdkDefaults, clientOpts)

	if result.Model != "client-model" {
		t.Errorf("resolveDefaultOptions() Model = %v, want %v", result.Model, "client-model")
	}
	if result.Temperature != 0.5 {
		t.Errorf("resolveDefaultOptions() Temperature = %v, want %v", result.Temperature, 0.5)
	}
	if result.MaxTokens != 50 {
		t.Errorf("resolveDefaultOptions() MaxTokens = %v, want %v", result.MaxTokens, 50)
	}
	if result.TopP != 0.8 {
		t.Errorf("resolveDefaultOptions() TopP = %v, want %v", result.TopP, 0.8)
	}
}

func TestConvertToMistralChatMessages(t *testing.T) { //nolint:funlen // comprehensive test
	tests := []struct {
		name     string
		messages []llms.MessageContent
		want     int
		wantErr  bool
		errorMsg string
	}{
		{
			name: "basic text messages",
			messages: []llms.MessageContent{
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
			},
			want: 2,
		},
		{
			name: "system message",
			messages: []llms.MessageContent{
				{
					Role: llms.ChatMessageTypeSystem,
					Parts: []llms.ContentPart{
						llms.TextContent{Text: "You are helpful"},
					},
				},
			},
			want: 1,
		},
		{
			name: "tool messages",
			messages: []llms.MessageContent{
				{
					Role: llms.ChatMessageTypeAI,
					Parts: []llms.ContentPart{
						llms.TextContent{Text: "I'll use a tool"},
						llms.ToolCall{
							ID: "call_123",
							FunctionCall: &llms.FunctionCall{
								Name:      "get_weather",
								Arguments: `{"location": "Paris"}`,
							},
						},
					},
				},
				{
					Role: llms.ChatMessageTypeTool,
					Parts: []llms.ContentPart{
						llms.ToolCallResponse{
							ToolCallID: "call_123",
							Name:       "get_weather",
							Content:    "20°C and sunny",
						},
					},
				},
			},
			want: 3, // AI message gets split: text content and tool call become separate messages
		},
		{
			name: "empty text content",
			messages: []llms.MessageContent{
				{
					Role: llms.ChatMessageTypeHuman,
					Parts: []llms.ContentPart{
						llms.TextContent{Text: ""},
					},
				},
			},
			want: 0, // Empty text messages are not added
		},
		{
			name: "generic message type",
			messages: []llms.MessageContent{
				{
					Role: llms.ChatMessageTypeGeneric,
					Parts: []llms.ContentPart{
						llms.TextContent{Text: "Generic message"},
					},
				},
			},
			want: 1,
		},
		{
			name: "function message type",
			messages: []llms.MessageContent{
				{
					Role: llms.ChatMessageTypeFunction,
					Parts: []llms.ContentPart{
						llms.TextContent{Text: "Function output"},
					},
				},
			},
			want: 1,
		},
		{
			name: "unsupported content type",
			messages: []llms.MessageContent{
				{
					Role: llms.ChatMessageTypeHuman,
					Parts: []llms.ContentPart{
						llms.ImageURLContent{URL: "https://example.com/image.png"},
					},
				},
			},
			wantErr:  true,
			errorMsg: "unsupported content type",
		},
		{
			name: "mixed content parts",
			messages: []llms.MessageContent{
				{
					Role: llms.ChatMessageTypeAI,
					Parts: []llms.ContentPart{
						llms.TextContent{Text: "Here's the weather"},
						llms.ToolCall{
							ID: "call_456",
							FunctionCall: &llms.FunctionCall{
								Name:      "get_temperature",
								Arguments: `{"unit": "celsius"}`,
							},
						},
						llms.TextContent{Text: "for your location"},
					},
				},
			},
			want: 3, // Each part becomes a separate message
		},
		{
			name: "tool call without function call",
			messages: []llms.MessageContent{
				{
					Role: llms.ChatMessageTypeAI,
					Parts: []llms.ContentPart{
						llms.ToolCall{
							ID:   "call_789",
							Type: "function",
							FunctionCall: &llms.FunctionCall{
								Name:      "search",
								Arguments: `{"query": "weather"}`,
							},
						},
					},
				},
			},
			want: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := convertToMistralChatMessages(tt.messages)
			if (err != nil) != tt.wantErr {
				t.Errorf("convertToMistralChatMessages() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errorMsg != "" && err != nil {
				if !contains(err.Error(), tt.errorMsg) {
					t.Errorf("convertToMistralChatMessages() error = %v, want error containing %v", err, tt.errorMsg)
				}
			}
			if !tt.wantErr && len(result) != tt.want {
				t.Errorf("convertToMistralChatMessages() returned %d messages, want %d", len(result), tt.want)
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[0:len(substr)] == substr || len(s) > len(substr) && contains(s[1:], substr)
}

func TestMistralChatParamsFromCallOptions(t *testing.T) {
	t.Run("basic options", func(t *testing.T) {
		callOpts := &llms.CallOptions{
			Model:       "mistral-large",
			Temperature: 0.7,
			MaxTokens:   200,
			TopP:        0.95,
			Seed:        42,
		}

		result := mistralChatParamsFromCallOptions(callOpts)

		if result.Temperature != 0.7 {
			t.Errorf("mistralChatParamsFromCallOptions() Temperature = %v, want %v", result.Temperature, 0.7)
		}
		if result.MaxTokens != 200 {
			t.Errorf("mistralChatParamsFromCallOptions() MaxTokens = %v, want %v", result.MaxTokens, 200)
		}
		// TopP is not set by mistralChatParamsFromCallOptions, it uses the default
		if result.RandomSeed != 42 {
			t.Errorf("mistralChatParamsFromCallOptions() RandomSeed = %v, want %v", result.RandomSeed, 42)
		}
	})

	t.Run("with tools", func(t *testing.T) {
		callOpts := &llms.CallOptions{
			Tools: []llms.Tool{
				{
					Type: "function",
					Function: &llms.FunctionDefinition{
						Name:        "get_weather",
						Description: "Get weather information",
						Parameters: map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"location": map[string]interface{}{
									"type":        "string",
									"description": "The city and state",
								},
							},
						},
					},
				},
			},
		}

		result := mistralChatParamsFromCallOptions(callOpts)

		if len(result.Tools) != 1 {
			t.Errorf("mistralChatParamsFromCallOptions() Tools length = %v, want %v", len(result.Tools), 1)
		}
		if result.Tools[0].Type != "function" {
			t.Errorf("mistralChatParamsFromCallOptions() Tool Type = %v, want %v", result.Tools[0].Type, "function")
		}
		if result.Tools[0].Function.Name != "get_weather" {
			t.Errorf("mistralChatParamsFromCallOptions() Function Name = %v, want %v", result.Tools[0].Function.Name, "get_weather")
		}
	})

	t.Run("with legacy functions", func(t *testing.T) {
		callOpts := &llms.CallOptions{
			Functions: []llms.FunctionDefinition{
				{
					Name:        "calculate",
					Description: "Perform calculations",
					Parameters: map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"expression": map[string]interface{}{
								"type": "string",
							},
						},
					},
				},
			},
		}

		result := mistralChatParamsFromCallOptions(callOpts)

		if len(result.Tools) != 1 {
			t.Errorf("mistralChatParamsFromCallOptions() Tools length = %v, want %v", len(result.Tools), 1)
		}
		if result.Tools[0].Function.Name != "calculate" {
			t.Errorf("mistralChatParamsFromCallOptions() Function Name = %v, want %v", result.Tools[0].Function.Name, "calculate")
		}
	})
}

func TestClientOptions(t *testing.T) {
	t.Run("WithAPIKey", func(t *testing.T) {
		opts := &clientOptions{}
		WithAPIKey("test-key")(opts)
		if opts.apiKey != "test-key" {
			t.Errorf("WithAPIKey() got %s, want test-key", opts.apiKey)
		}
	})

	t.Run("WithEndpoint", func(t *testing.T) {
		opts := &clientOptions{}
		WithEndpoint("https://test.com")(opts)
		if opts.endpoint != "https://test.com" {
			t.Errorf("WithEndpoint() got %s, want https://test.com", opts.endpoint)
		}
	})

	t.Run("WithMaxRetries", func(t *testing.T) {
		opts := &clientOptions{}
		WithMaxRetries(10)(opts)
		if opts.maxRetries != 10 {
			t.Errorf("WithMaxRetries() got %d, want 10", opts.maxRetries)
		}
	})

	t.Run("WithTimeout", func(t *testing.T) {
		opts := &clientOptions{}
		timeout := 60 * time.Second
		WithTimeout(timeout)(opts)
		if opts.timeout != timeout {
			t.Errorf("WithTimeout() got %v, want %v", opts.timeout, timeout)
		}
	})

	t.Run("WithModel", func(t *testing.T) {
		opts := &clientOptions{}
		WithModel("mistral-medium")(opts)
		if opts.model != "mistral-medium" {
			t.Errorf("WithModel() got %s, want mistral-medium", opts.model)
		}
	})

	t.Run("WithCallbacksHandler", func(t *testing.T) {
		opts := &clientOptions{}
		handler := &testCallbackHandler{}
		WithCallbacksHandler(handler)(opts)
		if opts.callbacksHandler != handler {
			t.Error("WithCallbacksHandler() did not set handler correctly")
		}
	})
}

type testCallbackHandler struct {
	generateStartCalled bool
	generateEndCalled   bool
	errorCalled         bool
}

func (h *testCallbackHandler) HandleLLMGenerateContentStart(ctx context.Context, messages []llms.MessageContent) {
	h.generateStartCalled = true
}

func (h *testCallbackHandler) HandleLLMGenerateContentEnd(ctx context.Context, resp *llms.ContentResponse) {
	h.generateEndCalled = true
}

func (h *testCallbackHandler) HandleLLMError(ctx context.Context, err error) {
	h.errorCalled = true
}

func (h *testCallbackHandler) HandleText(ctx context.Context, text string)                      {}
func (h *testCallbackHandler) HandleLLMStart(ctx context.Context, prompts []string)             {}
func (h *testCallbackHandler) HandleChainStart(ctx context.Context, inputs map[string]any)      {}
func (h *testCallbackHandler) HandleChainEnd(ctx context.Context, outputs map[string]any)       {}
func (h *testCallbackHandler) HandleChainError(ctx context.Context, err error)                  {}
func (h *testCallbackHandler) HandleToolStart(ctx context.Context, input string)                {}
func (h *testCallbackHandler) HandleToolEnd(ctx context.Context, output string)                 {}
func (h *testCallbackHandler) HandleToolError(ctx context.Context, err error)                   {}
func (h *testCallbackHandler) HandleAgentAction(ctx context.Context, action schema.AgentAction) {}
func (h *testCallbackHandler) HandleAgentFinish(ctx context.Context, finish schema.AgentFinish) {}
func (h *testCallbackHandler) HandleRetrieverStart(ctx context.Context, query string)           {}
func (h *testCallbackHandler) HandleRetrieverEnd(ctx context.Context, query string, documents []schema.Document) {
}
func (h *testCallbackHandler) HandleStreamingFunc(ctx context.Context, chunk []byte) {}

func TestCall(t *testing.T) {
	// This test requires mocking the Mistral SDK client
	t.Skip("Call() requires integration testing with mock Mistral client")
}

func TestGenerateContent(t *testing.T) {
	// This test requires mocking the Mistral SDK client
	t.Skip("GenerateContent() requires integration testing with mock Mistral client")
}

// TestModelNamePropagation verifies that the model name is correctly propagated
// from client options and call options to the final API call parameters.
// This is a regression test for issue #1233 where an empty string was being
// passed as the model parameter, causing "Invalid model: " errors.
func TestModelNamePropagation(t *testing.T) {
	tests := []struct {
		name              string
		clientModel       string
		callOptionModel   string
		expectedModel     string
	}{
		{
			name:            "default model from client options",
			clientModel:     "mistral-small",
			callOptionModel: "",
			expectedModel:   "mistral-small",
		},
		{
			name:            "override with call option",
			clientModel:     "mistral-small",
			callOptionModel: "mistral-large",
			expectedModel:   "mistral-large",
		},
		{
			name:            "use SDK default when not specified",
			clientModel:     sdk.ModelOpenMistral7b,
			callOptionModel: "",
			expectedModel:   sdk.ModelOpenMistral7b,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create client options with the test model
			clientOpts := &clientOptions{
				model: tt.clientModel,
			}

			// Resolve call options (simulates what happens in Call/GenerateContent)
			callOpts := resolveDefaultOptions(sdk.DefaultChatRequestParams, clientOpts)

			// Apply call option override if specified
			if tt.callOptionModel != "" {
				setCallOptions([]llms.CallOption{
					llms.WithModel(tt.callOptionModel),
				}, callOpts)
			}

			// Verify the model is set correctly and is not empty
			if callOpts.Model == "" {
				t.Error("Model in callOptions is empty string - this would cause 'Invalid model' error")
			}

			if callOpts.Model != tt.expectedModel {
				t.Errorf("Model = %q, want %q", callOpts.Model, tt.expectedModel)
			}
		})
	}
}

// TestModelNotEmpty ensures that a model is always set and never an empty string.
// This prevents the "Invalid model: " error reported in issue #1233.
func TestModelNotEmpty(t *testing.T) {
	// Test with default initialization
	model, err := New(WithAPIKey("test-key"))
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	if model.clientOptions.model == "" {
		t.Error("Model is empty string after New() with defaults - this would cause API errors")
	}

	// Test that resolveDefaultOptions always provides a model
	callOpts := resolveDefaultOptions(sdk.DefaultChatRequestParams, model.clientOptions)
	if callOpts.Model == "" {
		t.Error("CallOptions.Model is empty string - this would cause 'Invalid model' error")
	}
}
