package bedrockclient

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/document"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/llms/reasoning"
)

// Note: BedrockRuntimeClientInterface is defined in bedrockclient_conversion.go

// MockBedrockRuntimeClient is a mock implementation of bedrock runtime client
type MockBedrockRuntimeClient struct {
	mock.Mock
}

func (m *MockBedrockRuntimeClient) Converse(ctx context.Context, input *bedrockruntime.ConverseInput, opts ...func(*bedrockruntime.Options)) (*bedrockruntime.ConverseOutput, error) {
	args := m.Called(ctx, input, opts)
	return args.Get(0).(*bedrockruntime.ConverseOutput), args.Error(1)
}

func (m *MockBedrockRuntimeClient) ConverseStream(ctx context.Context, input *bedrockruntime.ConverseStreamInput, opts ...func(*bedrockruntime.Options)) (*bedrockruntime.ConverseStreamOutput, error) {
	args := m.Called(ctx, input, opts)
	return args.Get(0).(*bedrockruntime.ConverseStreamOutput), args.Error(1)
}

func TestNewConverseClient(t *testing.T) {
	mockClient := &MockBedrockRuntimeClient{}
	client := NewConverseClient(mockClient)

	assert.NotNil(t, client)
	assert.Equal(t, mockClient, client.client)
}

func TestConverseClient_BasicTextCompletion(t *testing.T) {
	mockClient := &MockBedrockRuntimeClient{}
	client := NewConverseClient(mockClient)

	// Setup expected response
	expectedResponse := &bedrockruntime.ConverseOutput{
		Output: &types.ConverseOutputMemberMessage{
			Value: types.Message{
				Role: types.ConversationRoleAssistant,
				Content: []types.ContentBlock{
					&types.ContentBlockMemberText{
						Value: "Hello! How can I help you today?",
					},
				},
			},
		},
	}

	mockClient.On("Converse", mock.Anything, mock.Anything, mock.Anything).Return(expectedResponse, nil)

	// Test input
	input := &ConverseInput{
		ModelID: "anthropic.claude-3-sonnet-20240229-v1:0",
		Messages: []Message{
			{
				Role:    llms.ChatMessageTypeHuman,
				Content: "Hello",
				Type:    "text",
			},
		},
		MaxTokens:   ptr(1000),
		Temperature: ptr(0.7),
	}

	// Execute
	response, err := client.CreateCompletionConverse(t.Context(), input)

	// Verify
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Len(t, response.Choices, 1)
	assert.Equal(t, "Hello! How can I help you today?", response.Choices[0].Content)

	mockClient.AssertExpectations(t)
}

func TestConverseClient_ToolCalling(t *testing.T) {
	mockClient := &MockBedrockRuntimeClient{}
	client := NewConverseClient(mockClient)

	// Create tool input document
	toolInput := map[string]any{
		"location": "New York",
	}

	// Setup expected response with tool call
	expectedResponse := &bedrockruntime.ConverseOutput{
		Output: &types.ConverseOutputMemberMessage{
			Value: types.Message{
				Role: types.ConversationRoleAssistant,
				Content: []types.ContentBlock{
					&types.ContentBlockMemberToolUse{
						Value: types.ToolUseBlock{
							ToolUseId: ptr("tool_123"),
							Name:      ptr("get_weather"),
							Input:     document.NewLazyDocument(toolInput),
						},
					},
				},
			},
		},
	}

	mockClient.On("Converse", mock.Anything, mock.Anything, mock.Anything).Return(expectedResponse, nil)

	// Test input with tools
	tools := []llms.Tool{
		{
			Type: "function",
			Function: &llms.FunctionDefinition{
				Name:        "get_weather",
				Description: "Get weather information",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"location": map[string]any{
							"type":        "string",
							"description": "The location",
						},
					},
					"required": []string{"location"},
				},
			},
		},
	}

	input := &ConverseInput{
		ModelID: "anthropic.claude-3-sonnet-20240229-v1:0",
		Messages: []Message{
			{
				Role:    llms.ChatMessageTypeHuman,
				Content: "What's the weather in New York?",
				Type:    "text",
			},
		},
		Tools: tools,
	}

	// Execute
	response, err := client.CreateCompletionConverse(t.Context(), input)

	// Verify
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Len(t, response.Choices, 1)
	assert.Len(t, response.Choices[0].ToolCalls, 1)
	assert.Equal(t, "tool_123", response.Choices[0].ToolCalls[0].ID)
	assert.Equal(t, "get_weather", response.Choices[0].ToolCalls[0].FunctionCall.Name)
	assert.Equal(t, `{"location":"New York"}`, response.Choices[0].ToolCalls[0].FunctionCall.Arguments)

	mockClient.AssertExpectations(t)
}

func TestConverseClient_SystemMessages(t *testing.T) {
	mockClient := &MockBedrockRuntimeClient{}
	client := NewConverseClient(mockClient)

	// Setup expected response
	expectedResponse := &bedrockruntime.ConverseOutput{
		Output: &types.ConverseOutputMemberMessage{
			Value: types.Message{
				Role: types.ConversationRoleAssistant,
				Content: []types.ContentBlock{
					&types.ContentBlockMemberText{
						Value: "I'll be helpful and concise.",
					},
				},
			},
		},
	}

	mockClient.On("Converse", mock.Anything, mock.Anything, mock.Anything).Return(expectedResponse, nil)

	// Test input with system message
	input := &ConverseInput{
		ModelID: "anthropic.claude-3-sonnet-20240229-v1:0",
		Messages: []Message{
			{
				Role:    llms.ChatMessageTypeSystem,
				Content: "Be helpful and concise.",
				Type:    "text",
			},
			{
				Role:    llms.ChatMessageTypeHuman,
				Content: "Hello",
				Type:    "text",
			},
		},
	}

	// Execute
	response, err := client.CreateCompletionConverse(t.Context(), input)

	// Verify
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Len(t, response.Choices, 1)
	assert.Equal(t, "I'll be helpful and concise.", response.Choices[0].Content)

	mockClient.AssertExpectations(t)
}

func TestConverseClient_ModelDetection(t *testing.T) {
	client := &ConverseClient{}

	// Test reasoning support
	assert.True(t, client.supportsReasoning("us.anthropic.claude-opus-4-20250514-v1:0"))
	assert.True(t, client.supportsReasoning("us.anthropic.claude-sonnet-4-20250514-v1:0"))
	assert.True(t, client.supportsReasoning("us.anthropic.claude-3-7-sonnet-20250219-v1:0"))
	assert.False(t, client.supportsReasoning("us.amazon.nova-pro-v1:0"))
	assert.False(t, client.supportsReasoning("anthropic.claude-3-sonnet-20240229-v1:0"))
}

func TestConverseClient_InferenceConfiguration(t *testing.T) {
	mockClient := &MockBedrockRuntimeClient{}
	client := NewConverseClient(mockClient)

	// Mock that captures the input to verify inference config
	var capturedInput *bedrockruntime.ConverseInput
	mockClient.On("Converse", mock.Anything, mock.MatchedBy(func(input *bedrockruntime.ConverseInput) bool {
		capturedInput = input
		return true
	}), mock.Anything).Return(&bedrockruntime.ConverseOutput{
		Output: &types.ConverseOutputMemberMessage{
			Value: types.Message{
				Role: types.ConversationRoleAssistant,
				Content: []types.ContentBlock{
					&types.ContentBlockMemberText{Value: "Test response"},
				},
			},
		},
	}, nil)

	// Test input with various inference parameters
	input := &ConverseInput{
		ModelID: "anthropic.claude-3-sonnet-20240229-v1:0",
		Messages: []Message{
			{
				Role:    llms.ChatMessageTypeHuman,
				Content: "Hello",
				Type:    "text",
			},
		},
		MaxTokens:     ptr(1500),
		Temperature:   ptr(0.8),
		TopP:          ptr(0.9),
		StopSequences: []string{"STOP"},
	}

	// Execute
	_, err := client.CreateCompletionConverse(t.Context(), input)

	// Verify
	assert.NoError(t, err)
	assert.NotNil(t, capturedInput.InferenceConfig)
	assert.Equal(t, int32(1500), *capturedInput.InferenceConfig.MaxTokens)
	assert.Equal(t, float32(0.8), *capturedInput.InferenceConfig.Temperature)
	assert.Equal(t, float32(0.9), *capturedInput.InferenceConfig.TopP)
	assert.Equal(t, []string{"STOP"}, capturedInput.InferenceConfig.StopSequences)

	mockClient.AssertExpectations(t)
}

func TestConverseClient_EmptyResponse(t *testing.T) {
	mockClient := &MockBedrockRuntimeClient{}
	client := NewConverseClient(mockClient)

	// Setup empty response
	expectedResponse := &bedrockruntime.ConverseOutput{
		Output: nil,
	}

	mockClient.On("Converse", mock.Anything, mock.Anything, mock.Anything).Return(expectedResponse, nil)

	input := &ConverseInput{
		ModelID: "anthropic.claude-3-sonnet-20240229-v1:0",
		Messages: []Message{
			{
				Role:    llms.ChatMessageTypeHuman,
				Content: "Hello",
				Type:    "text",
			},
		},
	}

	// Execute
	response, err := client.CreateCompletionConverse(t.Context(), input)

	// Verify
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Empty(t, response.Choices)

	mockClient.AssertExpectations(t)
}

// Helper function to create pointers
func ptr[T any](v T) *T {
	return &v
}

// TestAIMessageAccumulator tests the AI message accumulator
func TestAIMessageAccumulator(t *testing.T) {
	t.Run("empty accumulator", func(t *testing.T) {
		accum := &aiMessageAccumulator{}
		assert.True(t, accum.isEmpty())
	})

	t.Run("accumulate text only", func(t *testing.T) {
		accum := &aiMessageAccumulator{}
		accum.addTextContent("Hello", nil)

		assert.False(t, accum.isEmpty())
		msg := accum.build()
		assert.Equal(t, types.ConversationRoleAssistant, msg.Role)
		assert.Len(t, msg.Content, 1)

		textBlock, ok := msg.Content[0].(*types.ContentBlockMemberText)
		assert.True(t, ok)
		assert.Equal(t, "Hello", textBlock.Value)
	})

	t.Run("accumulate text with reasoning", func(t *testing.T) {
		accum := &aiMessageAccumulator{}
		reasoningContent := &reasoning.ContentReasoning{
			Content:   "thinking",
			Signature: []byte("sig"),
		}
		accum.addTextContent("Hello", reasoningContent)

		msg := accum.build()
		assert.Len(t, msg.Content, 2) // reasoning + text

		reasoningBlock, ok := msg.Content[0].(*types.ContentBlockMemberReasoningContent)
		assert.True(t, ok)
		assert.NotNil(t, reasoningBlock.Value)

		textBlock, ok := msg.Content[1].(*types.ContentBlockMemberText)
		assert.True(t, ok)
		assert.Equal(t, "Hello", textBlock.Value)
	})

	t.Run("accumulate multiple tool calls", func(t *testing.T) {
		accum := &aiMessageAccumulator{}
		toolCall1 := &ToolCall{
			ID:        "tool1",
			Name:      "func1",
			Arguments: map[string]any{"arg": "val1"},
		}
		toolCall2 := &ToolCall{
			ID:        "tool2",
			Name:      "func2",
			Arguments: map[string]any{"arg": "val2"},
		}

		err := accum.addToolUse(toolCall1)
		assert.NoError(t, err)
		err = accum.addToolUse(toolCall2)
		assert.NoError(t, err)

		assert.False(t, accum.isEmpty())
		msg := accum.build()
		assert.Len(t, msg.Content, 2)
	})

	t.Run("accumulate text and tool calls", func(t *testing.T) {
		accum := &aiMessageAccumulator{}
		reasoningContent := &reasoning.ContentReasoning{Content: "thinking"}
		accum.addTextContent("Hello", reasoningContent)

		toolCall := &ToolCall{
			ID:        "tool1",
			Name:      "func1",
			Arguments: map[string]any{},
		}
		err := accum.addToolUse(toolCall)
		assert.NoError(t, err)

		msg := accum.build()
		// Should have: reasoning, text, tool use
		assert.Len(t, msg.Content, 3)
	})

	t.Run("accumulate with cache control", func(t *testing.T) {
		accum := &aiMessageAccumulator{}
		accum.addTextContent("Hello", nil)
		accum.setCacheControl(&CacheControl{Type: "ephemeral", TTL: "1h"})

		msg := accum.build()
		// Should have: text, cache point
		assert.Len(t, msg.Content, 2)

		_, ok := msg.Content[1].(*types.ContentBlockMemberCachePoint)
		assert.True(t, ok)
	})

	t.Run("reset clears accumulator", func(t *testing.T) {
		accum := &aiMessageAccumulator{}
		accum.addTextContent("Hello", nil)
		assert.False(t, accum.isEmpty())

		accum.reset()
		assert.True(t, accum.isEmpty())
	})
}

// TestToolResultAccumulator tests the tool result accumulator
func TestToolResultAccumulator(t *testing.T) {
	t.Run("empty accumulator", func(t *testing.T) {
		accum := &toolResultAccumulator{}
		assert.True(t, accum.isEmpty())
	})

	t.Run("accumulate single result", func(t *testing.T) {
		accum := &toolResultAccumulator{}
		result := &ToolResult{
			ToolCallID: "tool1",
			ToolName:   "func1",
			Content:    "result1",
		}

		err := accum.addToolResult(result)
		assert.NoError(t, err)
		assert.False(t, accum.isEmpty())

		msg := accum.build()
		assert.Equal(t, types.ConversationRoleUser, msg.Role)
		assert.Len(t, msg.Content, 1)
	})

	t.Run("accumulate multiple results", func(t *testing.T) {
		accum := &toolResultAccumulator{}
		result1 := &ToolResult{ToolCallID: "tool1", ToolName: "func1", Content: "result1"}
		result2 := &ToolResult{ToolCallID: "tool2", ToolName: "func2", Content: "result2"}

		err := accum.addToolResult(result1)
		assert.NoError(t, err)
		err = accum.addToolResult(result2)
		assert.NoError(t, err)

		msg := accum.build()
		assert.Len(t, msg.Content, 2)
	})

	t.Run("reset clears accumulator", func(t *testing.T) {
		accum := &toolResultAccumulator{}
		accum.addToolResult(&ToolResult{ToolCallID: "t1", Content: "r1"})
		assert.False(t, accum.isEmpty())

		accum.reset()
		assert.True(t, accum.isEmpty())
	})
}

// TestConvertMessages_MultipleToolCallVariants tests all variants of message chains
func TestConvertMessages_MultipleToolCallVariants(t *testing.T) {
	client := &ConverseClient{}

	// Common test data
	toolCall1 := &ToolCall{ID: "tool1", Name: "func1", Arguments: map[string]any{"a": "1"}}
	toolCall2 := &ToolCall{ID: "tool2", Name: "func2", Arguments: map[string]any{"b": "2"}}
	result1 := &ToolResult{ToolCallID: "tool1", ToolName: "func1", Content: "result1"}
	result2 := &ToolResult{ToolCallID: "tool2", ToolName: "func2", Content: "result2"}
	reasoningContent := &reasoning.ContentReasoning{Content: "thinking"}

	tests := []struct {
		name             string
		messages         []Message
		expectedMsgCount int
		validate         func(*testing.T, []types.Message)
	}{
		{
			name: "variant1_content_separate_toolcalls_together_results_together",
			messages: []Message{
				{Role: llms.ChatMessageTypeHuman, Content: "query"},
				{Role: llms.ChatMessageTypeAI, Content: "response", Reasoning: reasoningContent},
				{Role: llms.ChatMessageTypeAI, ToolCall: toolCall1},
				{Role: llms.ChatMessageTypeAI, ToolCall: toolCall2},
				{Role: llms.ChatMessageTypeTool, ToolResult: result1},
				{Role: llms.ChatMessageTypeTool, ToolResult: result2},
			},
			expectedMsgCount: 3, // user + assistant (with text+tools) + user (with results)
			validate: func(t *testing.T, msgs []types.Message) {
				// Check assistant message has reasoning, text, and 2 tool uses
				assert.Equal(t, types.ConversationRoleAssistant, msgs[1].Role)
				assert.GreaterOrEqual(t, len(msgs[1].Content), 4) // reasoning + text + 2 tools

				// Check user message has 2 tool results
				assert.Equal(t, types.ConversationRoleUser, msgs[2].Role)
				assert.Len(t, msgs[2].Content, 2)
			},
		},
		{
			name: "variant2_content_separate_toolcalls_separate_results_together",
			messages: []Message{
				{Role: llms.ChatMessageTypeHuman, Content: "query"},
				{Role: llms.ChatMessageTypeAI, Content: "response", Reasoning: reasoningContent},
				{Role: llms.ChatMessageTypeAI, ToolCall: toolCall1},
				{Role: llms.ChatMessageTypeAI, ToolCall: toolCall2},
				{Role: llms.ChatMessageTypeTool, ToolResult: result1},
				{Role: llms.ChatMessageTypeTool, ToolResult: result2},
			},
			expectedMsgCount: 3,
			validate: func(t *testing.T, msgs []types.Message) {
				assert.Equal(t, types.ConversationRoleAssistant, msgs[1].Role)
				assert.Equal(t, types.ConversationRoleUser, msgs[2].Role)
			},
		},
		{
			name: "variant3_content_separate_toolcalls_separate_results_separate",
			messages: []Message{
				{Role: llms.ChatMessageTypeHuman, Content: "query"},
				{Role: llms.ChatMessageTypeAI, Content: "response", Reasoning: reasoningContent},
				{Role: llms.ChatMessageTypeAI, ToolCall: toolCall1},
				{Role: llms.ChatMessageTypeAI, ToolCall: toolCall2},
				{Role: llms.ChatMessageTypeTool, ToolResult: result1},
				{Role: llms.ChatMessageTypeTool, ToolResult: result2},
			},
			expectedMsgCount: 3,
			validate: func(t *testing.T, msgs []types.Message) {
				// All tool results should be combined into one user message
				assert.Equal(t, types.ConversationRoleUser, msgs[2].Role)
				assert.Len(t, msgs[2].Content, 2)
			},
		},
		{
			name: "variant4_content_with_toolcalls_together_results_together",
			messages: []Message{
				{Role: llms.ChatMessageTypeHuman, Content: "query"},
				{Role: llms.ChatMessageTypeAI, Content: "response", Reasoning: reasoningContent},
				{Role: llms.ChatMessageTypeAI, ToolCall: toolCall1},
				{Role: llms.ChatMessageTypeAI, ToolCall: toolCall2},
				{Role: llms.ChatMessageTypeTool, ToolResult: result1},
				{Role: llms.ChatMessageTypeTool, ToolResult: result2},
			},
			expectedMsgCount: 3,
			validate: func(t *testing.T, msgs []types.Message) {
				assert.Equal(t, types.ConversationRoleAssistant, msgs[1].Role)
				assert.Equal(t, types.ConversationRoleUser, msgs[2].Role)
			},
		},
		{
			name: "variant5_content_with_toolcall1_separate_toolcall2_results_separate",
			messages: []Message{
				{Role: llms.ChatMessageTypeHuman, Content: "query"},
				{Role: llms.ChatMessageTypeAI, Content: "response", Reasoning: reasoningContent},
				{Role: llms.ChatMessageTypeAI, ToolCall: toolCall1},
				{Role: llms.ChatMessageTypeAI, ToolCall: toolCall2},
				{Role: llms.ChatMessageTypeTool, ToolResult: result1},
				{Role: llms.ChatMessageTypeTool, ToolResult: result2},
			},
			expectedMsgCount: 3,
			validate: func(t *testing.T, msgs []types.Message) {
				assert.Equal(t, types.ConversationRoleAssistant, msgs[1].Role)
				assert.Equal(t, types.ConversationRoleUser, msgs[2].Role)
			},
		},
		{
			name: "variant6_content_with_toolcalls_together_results_separate",
			messages: []Message{
				{Role: llms.ChatMessageTypeHuman, Content: "query"},
				{Role: llms.ChatMessageTypeAI, Content: "response", Reasoning: reasoningContent},
				{Role: llms.ChatMessageTypeAI, ToolCall: toolCall1},
				{Role: llms.ChatMessageTypeAI, ToolCall: toolCall2},
				{Role: llms.ChatMessageTypeTool, ToolResult: result1},
				{Role: llms.ChatMessageTypeTool, ToolResult: result2},
			},
			expectedMsgCount: 3,
			validate: func(t *testing.T, msgs []types.Message) {
				assert.Equal(t, types.ConversationRoleAssistant, msgs[1].Role)
				assert.Equal(t, types.ConversationRoleUser, msgs[2].Role)
				// Both tool results should be in one user message
				assert.Len(t, msgs[2].Content, 2)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			messages, systemPrompts, err := client.convertMessages(tt.messages)

			assert.NoError(t, err)
			assert.Empty(t, systemPrompts)
			assert.Equal(t, tt.expectedMsgCount, len(messages))

			if tt.validate != nil {
				tt.validate(t, messages)
			}
		})
	}
}
