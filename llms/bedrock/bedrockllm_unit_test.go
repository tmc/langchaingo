package bedrock

import (
	"context"
	"slices"
	"testing"

	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/llms/bedrock/internal/bedrockclient"
	"github.com/vxcontrol/langchaingo/llms/streaming"
	"github.com/vxcontrol/langchaingo/schema"

	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
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
				WithModel(ModelAnthropicClaude37Sonnet),
			},
		},
		{
			name: "with custom model provider",
			opts: []Option{
				WithClient(&bedrockruntime.Client{}),
				WithModelProvider("anthropic"),
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
	ctx := t.Context()
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
		WithModel(ModelAnthropicClaude35Haiku)(opts)
		if opts.modelID != ModelAnthropicClaude35Haiku {
			t.Errorf("WithModel() got %s, want %s", opts.modelID, ModelAnthropicClaude35Haiku)
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

	t.Run("WithConverseAPI", func(t *testing.T) {
		opts := &options{}
		WithConverseAPI()(opts)
		if !opts.useConverseAPI {
			t.Error("WithConverseAPI() did not enable Converse API")
		}
	})

	t.Run("WithAutomaticCaching", func(t *testing.T) {
		opts := &options{}
		WithAutomaticCaching()(opts)
		if !opts.enableAutoCaching {
			t.Error("WithAutomaticCaching() did not enable automatic caching")
		}
	})
}

func TestModelConstants(t *testing.T) {
	// Test that some key model constants are defined
	models := []string{
		ModelAi21Jamba15LargeV1,
		ModelAi21Jamba15MiniV1,
		ModelAmazonNova2LiteV1,
		ModelAmazonNovaPremiereV1,
		ModelAmazonNovaProV1,
		ModelAmazonNovaLiteV1,
		ModelAmazonNovaMicroV1,
		ModelAnthropicClaudeOpus46,
		ModelAnthropicClaudeSonnet46,
		ModelAnthropicClaudeOpus45,
		ModelAnthropicClaudeHaiku45,
		ModelAnthropicClaudeSonnet45,
		ModelAnthropicClaudeOpus41,
		ModelAnthropicClaudeOpus4,
		ModelAnthropicClaudeSonnet4,
		ModelAnthropicClaude37Sonnet,
		ModelAnthropicClaude35Haiku,
		ModelCohereCommandRV1,
		ModelCohereCommandRPlusV1,
		ModelMetaLlama4MaverickInstructV1,
		ModelMetaLlama4ScoutInstructV1,
		ModelMetaLlama3370bInstructV1,
		ModelMetaLlama3211bInstructV1,
		ModelMetaLlama3290bInstructV1,
		ModelMetaLlama3170bInstructV1,
		ModelMetaLlama318bInstructV1,
		ModelMetaLlama370bInstructV1,
		ModelMetaLlama38bInstructV1,
		ModelDeepSeekR1V1,
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
	providers := []string{"ai21", "amazon", "anthropic", "cohere", "meta", "nova", "deepseek"}
	return slices.Contains(providers, bedrockclient.GetProvider(model))
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
func (h *testCallbackHandler) HandleStreamingFunc(ctx context.Context, chunk streaming.Chunk) {}

// TestSupportsCaching tests the supportsCaching method
func TestSupportsCaching(t *testing.T) {
	llm := &LLM{}

	tests := []struct {
		name     string
		modelID  string
		expected bool
	}{
		{
			name:     "Claude Opus 4.6 supports caching",
			modelID:  ModelAnthropicClaudeOpus46,
			expected: true,
		},
		{
			name:     "Claude Sonnet 4.6 supports caching",
			modelID:  ModelAnthropicClaudeSonnet46,
			expected: true,
		},
		{
			name:     "Claude Opus 4.5 supports caching",
			modelID:  ModelAnthropicClaudeOpus45,
			expected: true,
		},
		{
			name:     "Claude Haiku 4.5 supports caching",
			modelID:  ModelAnthropicClaudeHaiku45,
			expected: true,
		},
		{
			name:     "Claude Sonnet 4.5 supports caching",
			modelID:  ModelAnthropicClaudeSonnet45,
			expected: true,
		},
		{
			name:     "Claude Opus 4.1 supports caching",
			modelID:  ModelAnthropicClaudeOpus41,
			expected: true,
		},
		{
			name:     "Claude Opus 4 supports caching",
			modelID:  ModelAnthropicClaudeOpus4,
			expected: true,
		},
		{
			name:     "Claude Sonnet 4 supports caching",
			modelID:  ModelAnthropicClaudeSonnet4,
			expected: true,
		},
		{
			name:     "Claude 3.7 Sonnet does not support caching",
			modelID:  ModelAnthropicClaude37Sonnet,
			expected: false,
		},
		{
			name:     "Claude 3.5 Haiku does not support caching",
			modelID:  ModelAnthropicClaude35Haiku,
			expected: false,
		},
		{
			name:     "Amazon Nova does not support caching",
			modelID:  ModelAmazonNovaProV1,
			expected: false,
		},
		{
			name:     "Meta Llama does not support caching",
			modelID:  ModelMetaLlama3370bInstructV1,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := llm.supportsCaching(tt.modelID)
			if result != tt.expected {
				t.Errorf("supportsCaching(%s) = %v, want %v", tt.modelID, result, tt.expected)
			}
		})
	}
}

// TestApplyAutomaticCaching tests the applyAutomaticCaching function
func TestApplyAutomaticCaching(t *testing.T) {
	tests := []struct {
		name             string
		messages         []bedrockclient.Message
		expectedCacheIdx int
		expectNoCaching  bool
	}{
		{
			name: "cache applied to last assistant message",
			messages: []bedrockclient.Message{
				{Role: llms.ChatMessageTypeHuman, Type: "text", Content: "Hello"},
				{Role: llms.ChatMessageTypeAI, Type: "text", Content: "Hi there"},
				{Role: llms.ChatMessageTypeHuman, Type: "text", Content: "How are you?"},
			},
			expectedCacheIdx: 1,
		},
		{
			name: "cache applied to tool result before human message",
			messages: []bedrockclient.Message{
				{Role: llms.ChatMessageTypeHuman, Type: "text", Content: "What's the weather?"},
				{Role: llms.ChatMessageTypeAI, Type: "tool_use"},
				{Role: llms.ChatMessageTypeTool, Type: "tool_result", Content: "72F"},
				{Role: llms.ChatMessageTypeHuman, Type: "text", Content: "Thanks"},
			},
			expectedCacheIdx: 2,
		},
		{
			name: "no caching for single human message",
			messages: []bedrockclient.Message{
				{Role: llms.ChatMessageTypeHuman, Type: "text", Content: "Hello"},
			},
			expectNoCaching: true,
		},
		{
			name: "no caching when last message is system",
			messages: []bedrockclient.Message{
				{Role: llms.ChatMessageTypeSystem, Type: "text", Content: "You are helpful"},
			},
			expectNoCaching: true,
		},
		{
			name: "cache applied to assistant message before final human",
			messages: []bedrockclient.Message{
				{Role: llms.ChatMessageTypeSystem, Type: "text", Content: "System"},
				{Role: llms.ChatMessageTypeHuman, Type: "text", Content: "Hello"},
				{Role: llms.ChatMessageTypeAI, Type: "text", Content: "Hi"},
				{Role: llms.ChatMessageTypeHuman, Type: "text", Content: "Tell me more"},
			},
			expectedCacheIdx: 2,
		},
		{
			name:            "empty messages",
			messages:        []bedrockclient.Message{},
			expectNoCaching: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Make a copy to avoid modifying test data
			messages := make([]bedrockclient.Message, len(tt.messages))
			copy(messages, tt.messages)

			applyAutomaticCaching(messages)

			if tt.expectNoCaching {
				// Verify no cache control was added
				for i, msg := range messages {
					if msg.CacheControl != nil {
						t.Errorf("Expected no cache control, but found at index %d", i)
					}
				}
			} else {
				// Verify cache control was added to expected message
				if messages[tt.expectedCacheIdx].CacheControl == nil {
					t.Errorf("Expected cache control at index %d, but not found", tt.expectedCacheIdx)
				} else {
					if messages[tt.expectedCacheIdx].CacheControl.Type != "ephemeral" {
						t.Errorf("Expected cache type 'ephemeral', got '%s'", messages[tt.expectedCacheIdx].CacheControl.Type)
					}
					if messages[tt.expectedCacheIdx].CacheControl.TTL != "5m" {
						t.Errorf("Expected TTL '5m', got '%s'", messages[tt.expectedCacheIdx].CacheControl.TTL)
					}
				}

				// Verify no other messages have cache control
				for i := range messages {
					if i != tt.expectedCacheIdx && messages[i].CacheControl != nil {
						t.Errorf("Unexpected cache control at index %d (expected only at %d)", i, tt.expectedCacheIdx)
					}
				}
			}
		})
	}
}

// TestProcessMessagesWithCaching tests the processMessagesWithCaching function
func TestProcessMessagesWithCaching(t *testing.T) {
	tests := []struct {
		name         string
		messages     []llms.MessageContent
		autoCaching  bool
		expectCache  bool
		cacheAtIndex int
	}{
		{
			name: "automatic caching enabled",
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
				{
					Role: llms.ChatMessageTypeHuman,
					Parts: []llms.ContentPart{
						llms.TextContent{Text: "How are you?"},
					},
				},
			},
			autoCaching:  true,
			expectCache:  true,
			cacheAtIndex: 1,
		},
		{
			name: "automatic caching disabled",
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
			autoCaching: false,
			expectCache: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := processMessagesWithCaching(tt.messages, tt.autoCaching)
			if err != nil {
				t.Errorf("processMessagesWithCaching() error = %v", err)
				return
			}

			if tt.expectCache {
				if result[tt.cacheAtIndex].CacheControl == nil {
					t.Errorf("Expected cache control at index %d, but not found", tt.cacheAtIndex)
				}
			} else {
				for i, msg := range result {
					if msg.CacheControl != nil {
						t.Errorf("Unexpected cache control at index %d", i)
					}
				}
			}
		})
	}
}
