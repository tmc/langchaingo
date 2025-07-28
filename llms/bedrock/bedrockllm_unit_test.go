package bedrock

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/0xDezzy/langchaingo/llms"
	"github.com/0xDezzy/langchaingo/schema"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		opts    []Option
		wantErr bool
	}{
		{
			name: "with default options",
			opts: []Option{WithClient(&bedrockruntime.Client{})},
		},
		{
			name: "with custom model",
			opts: []Option{
				WithClient(&bedrockruntime.Client{}),
				WithModel(ModelAnthropicClaudeV3Sonnet),
			},
		},
		{
			name: "with callback handler",
			opts: []Option{
				WithClient(&bedrockruntime.Client{}),
				WithCallback(&testCallbackHandler{}),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			llm, err := New(tt.opts...)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && llm == nil {
				t.Error("New() returned nil LLM without error")
			}
		})
	}
}

func TestNewWithContext(t *testing.T) {
	ctx := context.Background()
	llm, err := NewWithContext(ctx, WithClient(&bedrockruntime.Client{}))
	if err != nil {
		t.Fatalf("NewWithContext() error: %v", err)
	}
	if llm == nil {
		t.Error("NewWithContext() returned nil LLM")
	}
}

func TestProcessMessages(t *testing.T) {
	tests := []struct {
		name     string
		messages []llms.MessageContent
		want     int
		wantErr  bool
	}{
		{
			name: "text messages",
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
						llms.TextContent{Text: "Hi there"},
					},
				},
			},
			want: 2,
		},
		{
			name: "binary content",
			messages: []llms.MessageContent{
				{
					Role: llms.ChatMessageTypeHuman,
					Parts: []llms.ContentPart{
						llms.BinaryContent{
							Data:     []byte("image"),
							MIMEType: "image/png",
						},
					},
				},
			},
			want: 1,
		},
		{
			name: "mixed content",
			messages: []llms.MessageContent{
				{
					Role: llms.ChatMessageTypeHuman,
					Parts: []llms.ContentPart{
						llms.TextContent{Text: "Look at this:"},
						llms.BinaryContent{
							Data:     []byte("image"),
							MIMEType: "image/jpeg",
						},
					},
				},
			},
			want: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := processMessages(tt.messages)
			if (err != nil) != tt.wantErr {
				t.Errorf("processMessages() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(result) != tt.want {
				t.Errorf("processMessages() returned %d messages, want %d", len(result), tt.want)
			}
		})
	}
}

func TestOptions(t *testing.T) {
	t.Run("WithModel", func(t *testing.T) {
		opts := &options{}
		WithModel(ModelAnthropicClaudeV3Haiku)(opts)
		if opts.modelID != ModelAnthropicClaudeV3Haiku {
			t.Errorf("WithModel() got %s, want %s", opts.modelID, ModelAnthropicClaudeV3Haiku)
		}
	})

	t.Run("WithClient", func(t *testing.T) {
		opts := &options{}
		client := &bedrockruntime.Client{}
		WithClient(client)(opts)
		if opts.client != client {
			t.Error("WithClient() did not set client correctly")
		}
	})

	t.Run("WithCallback", func(t *testing.T) {
		opts := &options{}
		handler := &testCallbackHandler{}
		WithCallback(handler)(opts)
		if opts.callbackHandler == nil {
			t.Error("WithCallback() did not set handler")
		}
	})
}

func TestModelConstants(t *testing.T) {
	// Test that some key model constants are defined
	models := []string{
		ModelAi21J2MidV1,
		ModelAi21J2UltraV1,
		ModelAmazonTitanTextLiteV1,
		ModelAmazonTitanTextExpressV1,
		ModelAnthropicClaudeV3Sonnet,
		ModelAnthropicClaudeV3Haiku,
		ModelCohereCommandTextV14,
		ModelMetaLlama270bChatV1,
	}

	for _, model := range models {
		if model == "" {
			t.Error("Model constant is empty")
		}
		if !containsProvider(model) {
			t.Errorf("Model %s does not contain a valid provider prefix", model)
		}
	}
}

func containsProvider(model string) bool {
	providers := []string{"ai21", "amazon", "anthropic", "cohere", "meta"}
	for _, provider := range providers {
		if len(model) > len(provider) && model[:len(provider)] == provider {
			return true
		}
	}
	return false
}

// Test helpers
type testCallbackHandler struct{}

func (h *testCallbackHandler) HandleLLMGenerateContentStart(ctx context.Context, messages []llms.MessageContent) {
}
func (h *testCallbackHandler) HandleLLMGenerateContentEnd(ctx context.Context, resp *llms.ContentResponse) {
}
func (h *testCallbackHandler) HandleLLMError(ctx context.Context, err error)                    {}
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
