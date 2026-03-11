package openai

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/llms/openai/internal/openaiclient"
	"github.com/vxcontrol/langchaingo/llms/reasoning"
)

// TestExtractToolParts tests the ExtractToolParts function with various content types
func TestExtractToolParts(t *testing.T) { //nolint:funlen
	t.Parallel()

	tests := []struct {
		name                     string
		multiContent             []llms.ContentPart
		expectedContentLen       int
		expectedToolCallsLen     int
		expectedToolResponsesLen int
		validateContent          func(t *testing.T, content []llms.ContentPart)
	}{
		{
			name: "empty text content with reasoning is ignored",
			multiContent: []llms.ContentPart{
				llms.TextContent{
					Text: "",
					Reasoning: &reasoning.ContentReasoning{
						Content: "This is reasoning content that should be ignored",
					},
				},
				llms.TextContent{
					Text: "This is normal text content",
				},
			},
			expectedContentLen:       1,
			expectedToolCallsLen:     0,
			expectedToolResponsesLen: 0,
			validateContent: func(t *testing.T, content []llms.ContentPart) {
				textContent, ok := content[0].(llms.TextContent)
				assert.True(t, ok, "Content should be TextContent type")
				assert.Equal(t, "This is normal text content", textContent.Text)
			},
		},
		{
			name: "multiple empty text content are filtered",
			multiContent: []llms.ContentPart{
				llms.TextContent{Text: "", Reasoning: &reasoning.ContentReasoning{Content: "reasoning 1"}},
				llms.TextContent{Text: "", Reasoning: &reasoning.ContentReasoning{Content: "reasoning 2"}},
				llms.TextContent{Text: "Valid text"},
			},
			expectedContentLen:       1,
			expectedToolCallsLen:     0,
			expectedToolResponsesLen: 0,
			validateContent: func(t *testing.T, content []llms.ContentPart) {
				textContent, ok := content[0].(llms.TextContent)
				assert.True(t, ok)
				assert.Equal(t, "Valid text", textContent.Text)
			},
		},
		{
			name: "all empty text content results in no content",
			multiContent: []llms.ContentPart{
				llms.TextContent{Text: "", Reasoning: &reasoning.ContentReasoning{Content: "reasoning only"}},
				llms.TextContent{Text: ""},
			},
			expectedContentLen:       0,
			expectedToolCallsLen:     0,
			expectedToolResponsesLen: 0,
		},
		{
			name: "mixed content types including empty text",
			multiContent: []llms.ContentPart{
				llms.TextContent{Text: "", Reasoning: &reasoning.ContentReasoning{Content: "reasoning"}},
				llms.ImageURLContent{URL: "https://example.com/image.png"},
				llms.TextContent{Text: "Some text"},
				llms.ToolCall{
					ID:   "call_123",
					Type: "function",
					FunctionCall: &llms.FunctionCall{
						Name:      "test_function",
						Arguments: "{}",
					},
				},
			},
			expectedContentLen:       2,
			expectedToolCallsLen:     1,
			expectedToolResponsesLen: 0,
			validateContent: func(t *testing.T, content []llms.ContentPart) {
				_, isImage := content[0].(llms.ImageURLContent)
				assert.True(t, isImage, "First content should be ImageURLContent")
				textContent, isText := content[1].(llms.TextContent)
				assert.True(t, isText, "Second content should be TextContent")
				assert.Equal(t, "Some text", textContent.Text)
			},
		},
		{
			name: "tool calls and responses are separated",
			multiContent: []llms.ContentPart{
				llms.ToolCall{ID: "call_1", Type: "function", FunctionCall: &llms.FunctionCall{Name: "func1"}},
				llms.ToolCallResponse{ToolCallID: "call_1", Name: "func1", Content: "result"},
				llms.TextContent{Text: "text"},
			},
			expectedContentLen:       1,
			expectedToolCallsLen:     1,
			expectedToolResponsesLen: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			msg := &ChatMessage{MultiContent: tt.multiContent}
			content, toolCalls, toolResponses := ExtractToolParts(msg)

			assert.Len(t, content, tt.expectedContentLen, "Unexpected content length")
			assert.Len(t, toolCalls, tt.expectedToolCallsLen, "Unexpected tool calls length")
			assert.Len(t, toolResponses, tt.expectedToolResponsesLen, "Unexpected tool responses length")

			if tt.validateContent != nil && len(content) > 0 {
				tt.validateContent(t, content)
			}
		})
	}
}

// TestSetMessageRole tests role conversion for different message types
func TestSetMessageRole(t *testing.T) {
	t.Parallel()

	llm := &LLM{}

	tests := []struct {
		name         string
		role         llms.ChatMessageType
		parts        []llms.ContentPart
		expectedRole string
		expectError  bool
	}{
		{
			name:         "system message",
			role:         llms.ChatMessageTypeSystem,
			parts:        []llms.ContentPart{llms.TextContent{Text: "System prompt"}},
			expectedRole: RoleSystem,
			expectError:  false,
		},
		{
			name:         "ai message",
			role:         llms.ChatMessageTypeAI,
			parts:        []llms.ContentPart{llms.TextContent{Text: "AI response"}},
			expectedRole: RoleAssistant,
			expectError:  false,
		},
		{
			name:         "human message",
			role:         llms.ChatMessageTypeHuman,
			parts:        []llms.ContentPart{llms.TextContent{Text: "User message"}},
			expectedRole: RoleUser,
			expectError:  false,
		},
		{
			name:         "generic message",
			role:         llms.ChatMessageTypeGeneric,
			parts:        []llms.ContentPart{llms.TextContent{Text: "Generic message"}},
			expectedRole: RoleUser,
			expectError:  false,
		},
		{
			name: "function message",
			role: llms.ChatMessageTypeFunction,
			parts: []llms.ContentPart{llms.ToolCallResponse{
				ToolCallID: "call_123",
				Name:       "function_name",
				Content:    "result",
			}},
			expectedRole: RoleFunction,
			expectError:  false,
		},
		{
			name: "tool message",
			role: llms.ChatMessageTypeTool,
			parts: []llms.ContentPart{llms.ToolCallResponse{
				ToolCallID: "call_456",
				Name:       "tool_name",
				Content:    "result",
			}},
			expectedRole: RoleTool,
			expectError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			msg := &ChatMessage{}
			mc := llms.MessageContent{Role: tt.role, Parts: tt.parts}

			err := llm.setMessageRole(msg, mc)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedRole, msg.Role)
			}
		})
	}
}

// TestHandleFunctionMessage tests function message handling
func TestHandleFunctionMessage(t *testing.T) {
	t.Parallel()

	llm := &LLM{}

	tests := []struct {
		name        string
		parts       []llms.ContentPart
		expectError bool
		validate    func(t *testing.T, msg *ChatMessage)
	}{
		{
			name: "valid function message",
			parts: []llms.ContentPart{llms.ToolCallResponse{
				ToolCallID: "call_123",
				Name:       "function_name",
				Content:    "result content",
			}},
			expectError: false,
			validate: func(t *testing.T, msg *ChatMessage) {
				assert.Equal(t, "call_123", msg.ToolCallID)
				assert.Equal(t, "function_name", msg.Name)
				assert.Equal(t, "result content", msg.Content)
			},
		},
		{
			name:        "too many parts",
			parts:       []llms.ContentPart{llms.TextContent{Text: "a"}, llms.TextContent{Text: "b"}},
			expectError: true,
		},
		{
			name:        "no parts",
			parts:       []llms.ContentPart{},
			expectError: true,
		},
		{
			name:        "wrong part type",
			parts:       []llms.ContentPart{llms.TextContent{Text: "text"}},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			msg := &ChatMessage{}
			mc := llms.MessageContent{Role: llms.ChatMessageTypeFunction, Parts: tt.parts}

			err := llm.handleFunctionMessage(msg, mc)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, msg)
				}
			}
		})
	}
}

// TestHandleToolMessage tests tool message validation
func TestHandleToolMessage(t *testing.T) {
	t.Parallel()

	llm := &LLM{}

	tests := []struct {
		name        string
		parts       []llms.ContentPart
		expectError bool
	}{
		{
			name: "valid tool message",
			parts: []llms.ContentPart{llms.ToolCallResponse{
				ToolCallID: "call_123",
				Name:       "tool_name",
				Content:    "result",
			}},
			expectError: false,
		},
		{
			name: "multiple tool responses",
			parts: []llms.ContentPart{
				llms.ToolCallResponse{ToolCallID: "call_1", Name: "tool_1", Content: "result1"},
				llms.ToolCallResponse{ToolCallID: "call_2", Name: "tool_2", Content: "result2"},
			},
			expectError: false,
		},
		{
			name: "tool response with text content",
			parts: []llms.ContentPart{
				llms.ToolCallResponse{ToolCallID: "call_1", Name: "tool_1", Content: "result"},
				llms.TextContent{Text: "additional text"},
			},
			expectError: false,
		},
		{
			name: "missing tool call ID",
			parts: []llms.ContentPart{llms.ToolCallResponse{
				ToolCallID: "",
				Name:       "tool_name",
				Content:    "result",
			}},
			expectError: true,
		},
		{
			name: "missing tool name",
			parts: []llms.ContentPart{llms.ToolCallResponse{
				ToolCallID: "call_123",
				Name:       "",
				Content:    "result",
			}},
			expectError: true,
		},
		{
			name:        "invalid part type",
			parts:       []llms.ContentPart{llms.ImageURLContent{URL: "http://example.com/img.png"}},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			mc := llms.MessageContent{Role: llms.ChatMessageTypeTool, Parts: tt.parts}

			err := llm.handleToolMessage(mc)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestProcessUsage tests usage statistics processing
func TestProcessUsage(t *testing.T) {
	t.Parallel()

	llm := &LLM{}

	usage := &openaiclient.ChatUsage{
		PromptTokens:     100,
		CompletionTokens: 50,
		TotalTokens:      150,
	}
	usage.PromptTokensDetails.CachedTokens = 20
	usage.PromptTokensDetails.CacheWriteTokens = 10
	usage.PromptTokensDetails.AudioTokens = 5
	usage.CompletionTokensDetails.ReasoningTokens = 15
	usage.CompletionTokensDetails.AudioTokens = 3
	usage.CompletionTokensDetails.AcceptedPredictionTokens = 8
	usage.CompletionTokensDetails.RejectedPredictionTokens = 2

	result := llm.processUsage(usage)

	assert.Equal(t, 50, result["CompletionTokens"])
	assert.Equal(t, 100, result["PromptTokens"]) // full prompt tokens
	assert.Equal(t, 150, result["TotalTokens"])
	assert.Equal(t, 15, result["ReasoningTokens"])
	assert.Equal(t, 20, result["PromptCachedTokens"])
	assert.Equal(t, 20, result["CacheReadInputTokens"])
	assert.Equal(t, 10, result["CacheCreationInputTokens"]) // CacheWriteTokens
	assert.Equal(t, 5, result["PromptAudioTokens"])
	assert.Equal(t, 3, result["CompletionAudioTokens"])
	assert.Equal(t, 15, result["CompletionReasoningTokens"])
	assert.Equal(t, 8, result["CompletionAcceptedPredictionTokens"])
	assert.Equal(t, 2, result["CompletionRejectedPredictionTokens"])
}

// TestTokenUsageMapping_OpenAI tests correct token usage mapping for OpenAI provider
func TestTokenUsageMapping_OpenAI(t *testing.T) { //nolint:funlen
	t.Parallel()

	llm := &LLM{}

	tests := []struct {
		name                     string
		promptTokens             int
		cachedTokens             int
		cacheWriteTokens         int
		completionTokens         int
		expectedPromptTokens     int
		expectedCacheRead        int
		expectedCacheCreation    int
		expectedCompletionTokens int
		expectedTotalTokens      int
	}{
		{
			name:                     "first request without cache",
			promptTokens:             2619,
			cachedTokens:             0,
			cacheWriteTokens:         0,
			completionTokens:         149,
			expectedPromptTokens:     2619,
			expectedCacheRead:        0,
			expectedCacheCreation:    0,
			expectedCompletionTokens: 149,
			expectedTotalTokens:      2768,
		},
		{
			name:                     "subsequent request with cache hit",
			promptTokens:             2619,
			cachedTokens:             2048,
			cacheWriteTokens:         0,
			completionTokens:         85,
			expectedPromptTokens:     2619,
			expectedCacheRead:        2048,
			expectedCacheCreation:    0,
			expectedCompletionTokens: 85,
			expectedTotalTokens:      2704,
		},
		{
			name:                     "request with cache write",
			promptTokens:             5000,
			cachedTokens:             0,
			cacheWriteTokens:         4500,
			completionTokens:         200,
			expectedPromptTokens:     5000,
			expectedCacheRead:        0,
			expectedCacheCreation:    4500,
			expectedCompletionTokens: 200,
			expectedTotalTokens:      5200,
		},
		{
			name:                     "mixed scenario",
			promptTokens:             3000,
			cachedTokens:             1500,
			cacheWriteTokens:         800,
			completionTokens:         120,
			expectedPromptTokens:     3000,
			expectedCacheRead:        1500,
			expectedCacheCreation:    800,
			expectedCompletionTokens: 120,
			expectedTotalTokens:      3120,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			usage := &openaiclient.ChatUsage{
				PromptTokens:     tt.promptTokens,
				CompletionTokens: tt.completionTokens,
				TotalTokens:      tt.promptTokens + tt.completionTokens,
			}
			usage.PromptTokensDetails.CachedTokens = tt.cachedTokens
			usage.PromptTokensDetails.CacheWriteTokens = tt.cacheWriteTokens

			result := llm.processUsage(usage)

			// Verify mapped values
			assert.Equal(t, tt.expectedPromptTokens, result["PromptTokens"], "PromptTokens mismatch")
			assert.Equal(t, tt.expectedCacheRead, result["CacheReadInputTokens"], "CacheReadInputTokens mismatch")
			assert.Equal(t, tt.expectedCacheCreation, result["CacheCreationInputTokens"], "CacheCreationInputTokens mismatch")
			assert.Equal(t, tt.expectedCompletionTokens, result["CompletionTokens"], "CompletionTokens mismatch")
			assert.Equal(t, tt.expectedTotalTokens, result["TotalTokens"], "TotalTokens mismatch")

			// Verify client-side cost calculation logic
			// Client formula: input = max(PromptTokens - CacheRead, 0)
			promptTokens := result["PromptTokens"].(int)
			cacheRead := result["CacheReadInputTokens"].(int)
			cacheWrite := result["CacheCreationInputTokens"].(int)

			uncachedTokens := max(promptTokens-cacheRead, 0)

			// Expected: uncached tokens should equal promptTokens - cachedTokens
			expectedUncached := tt.promptTokens - tt.cachedTokens
			assert.Equal(t, expectedUncached, uncachedTokens, "Uncached tokens calculation mismatch")

			// For OpenAI pricing: uncached * basePrice + cacheRead * cacheReadPrice
			// OpenAI doesn't charge extra for cache writes
			basePrice := 2.5 / 1e6       // $2.5 per 1M tokens
			cacheReadPrice := 1.25 / 1e6 // $1.25 per 1M tokens (50% discount)

			expectedCost := float64(uncachedTokens)*basePrice +
				float64(cacheRead)*cacheReadPrice

			actualCost := float64(uncachedTokens)*basePrice +
				float64(cacheRead)*cacheReadPrice

			assert.InDelta(t, expectedCost, actualCost, 0.000001, "Cost calculation mismatch")

			// Verify CacheWrite is not used in OpenAI pricing
			assert.Equal(t, tt.cacheWriteTokens, cacheWrite, "CacheWrite should be stored but not used in pricing")
		})
	}
}

// TestProcessReasoning tests reasoning content processing
func TestProcessReasoning(t *testing.T) {
	t.Parallel()

	llm := &LLM{}

	tests := []struct {
		name             string
		reasoningContent string
		expectNil        bool
	}{
		{
			name:             "with reasoning content",
			reasoningContent: "This is the reasoning process",
			expectNil:        false,
		},
		{
			name:             "empty reasoning content",
			reasoningContent: "",
			expectNil:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := llm.processReasoning(tt.reasoningContent)

			if tt.expectNil {
				assert.Nil(t, result)
			} else {
				require.NotNil(t, result)
				assert.Equal(t, tt.reasoningContent, result.Content)
				assert.Nil(t, result.Signature)
			}
		})
	}
}

// TestProcessToolCalls tests tool calls processing in response
func TestProcessToolCalls(t *testing.T) {
	t.Parallel()

	llm := &LLM{}

	tests := []struct {
		name              string
		finishReason      openaiclient.FinishReason
		functionCall      *openaiclient.FunctionCall
		toolCalls         []openaiclient.ToolCall
		expectedFuncCall  bool
		expectedToolCalls int
	}{
		{
			name:         "legacy function call",
			finishReason: "function_call",
			functionCall: &openaiclient.FunctionCall{
				Name:      "my_function",
				Arguments: `{"arg": "value"}`,
			},
			expectedFuncCall:  true,
			expectedToolCalls: 0,
		},
		{
			name:         "modern tool calls",
			finishReason: "tool_calls",
			toolCalls: []openaiclient.ToolCall{
				{
					ID:   "call_1",
					Type: "function",
					Function: openaiclient.ToolFunction{
						Name:      "tool_1",
						Arguments: `{"arg1": "value1"}`,
					},
				},
				{
					ID:   "call_2",
					Type: "function",
					Function: openaiclient.ToolFunction{
						Name:      "tool_2",
						Arguments: `{"arg2": "value2"}`,
					},
				},
			},
			expectedFuncCall:  true, // should populate legacy field
			expectedToolCalls: 2,
		},
		{
			name:              "no tool calls",
			finishReason:      "stop",
			expectedFuncCall:  false,
			expectedToolCalls: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			choice := &llms.ContentChoice{}
			c := &openaiclient.ChatCompletionChoice{
				FinishReason: tt.finishReason,
				Message: openaiclient.ChatMessage{
					FunctionCall: tt.functionCall,
					ToolCalls:    tt.toolCalls,
				},
			}

			llm.processToolCalls(choice, c)

			if tt.expectedFuncCall {
				assert.NotNil(t, choice.FuncCall)
			} else {
				assert.Nil(t, choice.FuncCall)
			}
			assert.Len(t, choice.ToolCalls, tt.expectedToolCalls)

			// Verify legacy field is populated with first tool call
			if tt.expectedToolCalls > 0 {
				assert.Equal(t, choice.ToolCalls[0].FunctionCall, choice.FuncCall)
			}
		})
	}
}

// TestToolConversions tests tool conversion functions
func TestToolConversions(t *testing.T) {
	t.Parallel()

	t.Run("toolFromTool", func(t *testing.T) {
		t.Parallel()

		tool := llms.Tool{
			Type: "function",
			Function: &llms.FunctionDefinition{
				Name:        "test_function",
				Description: "A test function",
				Parameters:  map[string]any{"type": "object"},
				Strict:      true,
			},
		}

		result, err := toolFromTool(tool)
		require.NoError(t, err)
		assert.Equal(t, openaiclient.ToolTypeFunction, result.Type)
		assert.Equal(t, "test_function", result.Function.Name)
		assert.Equal(t, "A test function", result.Function.Description)
		assert.True(t, result.Function.Strict)
	})

	t.Run("toolFromTool invalid type", func(t *testing.T) {
		t.Parallel()

		tool := llms.Tool{Type: "invalid_type"}
		_, err := toolFromTool(tool)
		assert.Error(t, err)
	})

	t.Run("toolCallFromToolCall", func(t *testing.T) {
		t.Parallel()

		tc := llms.ToolCall{
			ID:   "call_123",
			Type: "function",
			FunctionCall: &llms.FunctionCall{
				Name:      "my_func",
				Arguments: `{"key": "value"}`,
			},
		}

		result := toolCallFromToolCall(tc)
		assert.Equal(t, "call_123", result.ID)
		assert.Equal(t, openaiclient.ToolType("function"), result.Type)
		assert.Equal(t, "my_func", result.Function.Name)
		assert.Equal(t, `{"key": "value"}`, result.Function.Arguments)
	})

	t.Run("toolCallsFromToolCalls", func(t *testing.T) {
		t.Parallel()

		tcs := []llms.ToolCall{
			{
				ID:           "call_1",
				Type:         "function",
				FunctionCall: &llms.FunctionCall{Name: "func1", Arguments: "{}"},
			},
			{
				ID:           "call_2",
				Type:         "function",
				FunctionCall: &llms.FunctionCall{Name: "func2", Arguments: "{}"},
			},
		}

		results := toolCallsFromToolCalls(tcs)
		assert.Len(t, results, 2)
		assert.Equal(t, "call_1", results[0].ID)
		assert.Equal(t, "call_2", results[1].ID)
	})
}

// TestConvertMessages tests message conversion with various scenarios
func TestConvertMessages(t *testing.T) { //nolint:funlen
	t.Parallel()

	llm := &LLM{}

	tests := []struct {
		name          string
		messages      []llms.MessageContent
		expectedCount int
		expectError   bool
		validateFirst func(t *testing.T, msg *ChatMessage)
	}{
		{
			name: "simple text messages",
			messages: []llms.MessageContent{
				{
					Role:  llms.ChatMessageTypeSystem,
					Parts: []llms.ContentPart{llms.TextContent{Text: "You are helpful"}},
				},
				{
					Role:  llms.ChatMessageTypeHuman,
					Parts: []llms.ContentPart{llms.TextContent{Text: "Hello"}},
				},
			},
			expectedCount: 2,
			expectError:   false,
			validateFirst: func(t *testing.T, msg *ChatMessage) {
				assert.Equal(t, RoleSystem, msg.Role)
			},
		},
		{
			name: "message with empty text content is filtered",
			messages: []llms.MessageContent{
				{
					Role: llms.ChatMessageTypeHuman,
					Parts: []llms.ContentPart{
						llms.TextContent{Text: "", Reasoning: &reasoning.ContentReasoning{Content: "thinking"}},
						llms.TextContent{Text: "Real message"},
					},
				},
			},
			expectedCount: 1,
			expectError:   false,
			validateFirst: func(t *testing.T, msg *ChatMessage) {
				assert.Len(t, msg.MultiContent, 1)
			},
		},
		{
			name: "message with only empty content is skipped",
			messages: []llms.MessageContent{
				{
					Role:  llms.ChatMessageTypeHuman,
					Parts: []llms.ContentPart{llms.TextContent{Text: ""}},
				},
				{
					Role:  llms.ChatMessageTypeHuman,
					Parts: []llms.ContentPart{llms.TextContent{Text: "Valid"}},
				},
			},
			expectedCount: 1,
			expectError:   false,
		},
		{
			name: "tool call and response separation",
			messages: []llms.MessageContent{
				{
					Role: llms.ChatMessageTypeAI,
					Parts: []llms.ContentPart{
						llms.ToolCall{
							ID:           "call_1",
							Type:         "function",
							FunctionCall: &llms.FunctionCall{Name: "func", Arguments: "{}"},
						},
						llms.ToolCallResponse{
							ToolCallID: "call_1",
							Name:       "func",
							Content:    "result",
						},
					},
				},
			},
			expectedCount: 2, // One for assistant with tool call, one for tool response
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result, err := llm.convertMessages(tt.messages)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Len(t, result, tt.expectedCount)
				if tt.validateFirst != nil && len(result) > 0 {
					tt.validateFirst(t, result[0])
				}
			}
		})
	}
}

// TestErrorMapping tests the error mapping functionality
func TestErrorMapping(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		inputError   error
		expectedCode llms.ErrorCode
	}{
		{
			name:         "nil error",
			inputError:   nil,
			expectedCode: "",
		},
		{
			name:         "authentication error",
			inputError:   fmt.Errorf("incorrect api key provided"),
			expectedCode: llms.ErrCodeAuthentication,
		},
		{
			name:         "rate limit error",
			inputError:   fmt.Errorf("rate limit exceeded for requests"),
			expectedCode: llms.ErrCodeRateLimit,
		},
		{
			name:         "model not found",
			inputError:   fmt.Errorf("model not found: gpt-5"),
			expectedCode: llms.ErrCodeResourceNotFound,
		},
		{
			name:         "context length error",
			inputError:   fmt.Errorf("maximum context length exceeded"),
			expectedCode: llms.ErrCodeTokenLimit,
		},
		{
			name:         "content filter error",
			inputError:   fmt.Errorf("content filtering triggered"),
			expectedCode: llms.ErrCodeContentFilter,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := MapError(tt.inputError)

			if tt.inputError == nil {
				assert.Nil(t, result)
			} else {
				var llmErr *llms.Error
				require.ErrorAs(t, result, &llmErr, "Expected *llms.Error type")
				assert.Equal(t, tt.expectedCode, llmErr.Code)
				assert.Equal(t, "openai", llmErr.Provider)
			}
		})
	}
}

// TestCreateChatRequest_ReasoningModelTemperature tests temperature adjustment for reasoning models
func TestCreateChatRequest_ReasoningModelTemperature(t *testing.T) {
	t.Parallel()

	// Create a minimal client for testing
	client, err := openaiclient.New("fake-token", "", "", "", openaiclient.APITypeOpenAI, "", nil, "", nil, false, false, false)
	require.NoError(t, err)

	llm := &LLM{client: client}

	tests := []struct {
		name                string
		model               string
		temperature         float64
		expectedTemperature float64
	}{
		{
			name:                "reasoning model o1-preview with non-1.0 temperature",
			model:               "o1-preview",
			temperature:         0.7,
			expectedTemperature: 1.0,
		},
		{
			name:                "reasoning model o1-mini with zero temperature",
			model:               "o1-mini",
			temperature:         0.0,
			expectedTemperature: 1.0,
		},
		{
			name:                "reasoning model o3-mini with temperature 0.5",
			model:               "o3-mini",
			temperature:         0.5,
			expectedTemperature: 1.0,
		},
		{
			name:                "reasoning model with temperature already 1.0",
			model:               "o1-preview",
			temperature:         1.0,
			expectedTemperature: 1.0,
		},
		{
			name:                "non-reasoning model gpt-4 preserves temperature",
			model:               "gpt-4",
			temperature:         0.7,
			expectedTemperature: 0.7,
		},
		{
			name:                "non-reasoning model gpt-4o-mini preserves zero temperature",
			model:               "gpt-4o-mini",
			temperature:         0.0,
			expectedTemperature: 0.0,
		},
		{
			name:                "deepseek-r1 reasoning model adjusts temperature",
			model:               "deepseek-r1",
			temperature:         0.8,
			expectedTemperature: 1.0,
		},
		{
			name:                "gemini-2.5-flash reasoning model adjusts temperature",
			model:               "gemini-2.5-flash-thinking-exp",
			temperature:         0.3,
			expectedTemperature: 1.0,
		},
		{
			name:                "claude-3.7-sonnet reasoning model adjusts temperature",
			model:               "claude-3.7-sonnet",
			temperature:         0.9,
			expectedTemperature: 1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			opts := llms.CallOptions{
				Model:       &tt.model,
				Temperature: &tt.temperature,
			}

			req, err := llm.createChatRequest([]*ChatMessage{}, opts)
			require.NoError(t, err)
			require.NotNil(t, req.Temperature)
			if req.Temperature != nil {
				assert.Equal(t, tt.expectedTemperature, *req.Temperature,
					"Temperature should be %v for model %s", tt.expectedTemperature, tt.model)
			}
		})
	}
}

func TestExtraBody_Integration(t *testing.T) { //nolint:funlen
	t.Parallel()

	var receivedRequest map[string]any

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &receivedRequest)

		response := map[string]any{
			"id": "test-id",
			"choices": []map[string]any{
				{
					"index": 0,
					"message": map[string]any{
						"role":    "assistant",
						"content": "test response",
					},
					"finish_reason": "stop",
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	llm, err := New(
		WithToken("test-token"),
		WithBaseURL(server.URL),
		WithModel("test-model"),
	)
	require.NoError(t, err)

	tests := []struct {
		name      string
		extraBody map[string]any
		checkFunc func(t *testing.T, req map[string]any)
	}{
		{
			name: "simple extra fields",
			extraBody: map[string]any{
				"enable_thinking": false,
				"top_k":           20,
			},
			checkFunc: func(t *testing.T, req map[string]any) {
				assert.Equal(t, false, req["enable_thinking"])
				assert.Equal(t, float64(20), req["top_k"])
			},
		},
		{
			name: "nested extra fields",
			extraBody: map[string]any{
				"chat_template_kwargs": map[string]any{
					"enable_thinking": false,
				},
			},
			checkFunc: func(t *testing.T, req map[string]any) {
				kwargs, ok := req["chat_template_kwargs"].(map[string]any)
				require.True(t, ok)
				assert.Equal(t, false, kwargs["enable_thinking"])
			},
		},
		{
			name: "extra body overrides standard field",
			extraBody: map[string]any{
				"temperature": 0.9,
			},
			checkFunc: func(t *testing.T, req map[string]any) {
				assert.Equal(t, 0.9, req["temperature"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			receivedRequest = nil

			messages := []llms.MessageContent{
				{
					Role: llms.ChatMessageTypeHuman,
					Parts: []llms.ContentPart{
						llms.TextContent{Text: "test message"},
					},
				},
			}

			_, err := llm.GenerateContent(t.Context(), messages,
				WithExtraBody(tt.extraBody),
				llms.WithTemperature(0.7),
			)
			require.NoError(t, err)
			require.NotNil(t, receivedRequest)

			tt.checkFunc(t, receivedRequest)
		})
	}
}
