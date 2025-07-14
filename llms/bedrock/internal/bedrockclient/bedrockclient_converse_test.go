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
	response, err := client.CreateCompletionConverse(context.Background(), input)

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
	response, err := client.CreateCompletionConverse(context.Background(), input)

	// Verify
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Len(t, response.Choices, 1)
	assert.Len(t, response.Choices[0].ToolCalls, 1)
	assert.Equal(t, "tool_123", response.Choices[0].ToolCalls[0].ID)
	assert.Equal(t, "get_weather", response.Choices[0].ToolCalls[0].FunctionCall.Name)

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
	response, err := client.CreateCompletionConverse(context.Background(), input)

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
	_, err := client.CreateCompletionConverse(context.Background(), input)

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
	response, err := client.CreateCompletionConverse(context.Background(), input)

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
