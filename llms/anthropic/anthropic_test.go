package anthropic_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/llms/anthropic"
	"github.com/vxcontrol/langchaingo/llms/streaming"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAnthropic_GenerateContent(t *testing.T) {
	t.Parallel()

	llm := newHTTPRRClient(t, anthropic.WithModel("claude-sonnet-4-5"))

	parts := []llms.ContentPart{
		llms.TextContent{Text: "How many feet are in a nautical mile?"},
	}
	content := []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: parts,
		},
	}

	resp, err := llm.GenerateContent(t.Context(), content)
	require.NoError(t, err)

	assert.NotEmpty(t, resp.Choices)
	c1 := resp.Choices[0]
	assert.Regexp(t, "feet", strings.ToLower(c1.Content))
}

func TestAnthropic_GenerateContentWithTool(t *testing.T) {
	t.Parallel()

	availableTools := []llms.Tool{
		{
			Type: "function",
			Function: &llms.FunctionDefinition{
				Name:        "getCurrentWeather",
				Description: "Get the current weather in a given location",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"location": map[string]any{
							"type":        "string",
							"description": "The city and state, e.g. San Francisco, CA",
						},
						"unit": map[string]any{
							"type": "string",
							"enum": []string{"fahrenheit", "celsius"},
						},
					},
					"required": []string{"location"},
				},
			},
		},
	}

	llm := newHTTPRRClient(t, anthropic.WithModel("claude-haiku-4-5"))

	contents := []llms.MessageContent{ //nolint:prealloc
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextContent{Text: "What is the weather in Boston?"}},
		},
	}

	// Ask a questions about the weather and let the model know that the tool is available
	resp, err := llm.GenerateContent(t.Context(), contents, llms.WithTools(availableTools))
	require.NoError(t, err)

	// Expect a tool call in the response
	require.NotEmpty(t, resp.Choices)
	choice := resp.Choices[0]
	toolCall := choice.ToolCalls[0]
	assert.Equal(t, "getCurrentWeather", toolCall.FunctionCall.Name)

	// Append tool_use to contents
	assistantResponse := llms.MessageContent{
		Role: llms.ChatMessageTypeAI,
		Parts: []llms.ContentPart{
			llms.ToolCall{
				ID:   toolCall.ID,
				Type: toolCall.Type,
				FunctionCall: &llms.FunctionCall{
					Name:      toolCall.FunctionCall.Name,
					Arguments: toolCall.FunctionCall.Arguments,
				},
			},
		},
	}
	contents = append(contents, assistantResponse)

	// Call the tool
	currentWeather := `{"Boston","72 and sunny"}`

	// Append weather info to content
	weatherCallResponse := llms.MessageContent{
		Role: llms.ChatMessageTypeTool,
		Parts: []llms.ContentPart{
			llms.ToolCallResponse{
				ToolCallID: toolCall.ID,
				Name:       toolCall.FunctionCall.Name,
				Content:    currentWeather,
			},
		},
	}
	contents = append(contents, weatherCallResponse)

	// Generate answer with the tool response
	resp, err = llm.GenerateContent(t.Context(), contents, llms.WithTools(availableTools))
	require.NoError(t, err)

	require.NotEmpty(t, resp.Choices)
	choice = resp.Choices[0]
	assert.Regexp(t, "72", choice.Content)
}

func TestAnthropic_GenerateContentWithStreaming(t *testing.T) {
	t.Parallel()

	llm := newHTTPRRClient(t, anthropic.WithModel("claude-haiku-4-5"))

	content := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextContent{Text: "Count from 1 to 5"},
			},
		},
	}

	var (
		streamedChunks []string
		streamedDone   bool
	)

	streamingFunc := func(_ context.Context, chunk streaming.Chunk) error {
		switch chunk.Type { //nolint:exhaustive
		case streaming.ChunkTypeText:
			streamedChunks = append(streamedChunks, chunk.Content)
		case streaming.ChunkTypeDone:
			streamedDone = true
		default:
			// skip other chunks
		}
		return nil
	}

	resp, err := llm.GenerateContent(t.Context(), content, llms.WithStreamingFunc(streamingFunc))
	require.NoError(t, err)

	assert.True(t, streamedDone)
	assert.Greater(t, len(streamedChunks), 0)

	fullResponse := strings.Join(streamedChunks, "")

	assert.NotEmpty(t, resp.Choices)
	choice := resp.Choices[0]
	assert.Equal(t, fullResponse, choice.Content)

	for i := 1; i <= 5; i++ {
		assert.Contains(t, fullResponse, string(rune('0'+i)))
	}
}

//nolint:funlen,cyclop
func TestAnthropic_GenerateContentWithToolAndStreaming(t *testing.T) {
	t.Parallel()

	availableTools := []llms.Tool{
		{
			Type: "function",
			Function: &llms.FunctionDefinition{
				Name:        "calculateTip",
				Description: "Calculate tip amount based on bill total and percentage",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"billAmount": map[string]any{
							"type":        "number",
							"description": "The total bill amount",
						},
						"tipPercentage": map[string]any{
							"type":        "number",
							"description": "The tip percentage",
						},
					},
					"required": []string{"billAmount", "tipPercentage"},
				},
			},
		},
		{
			Type: "function",
			Function: &llms.FunctionDefinition{
				Name:        "getWaiterName",
				Description: "Get the name of the waiter",
				Parameters: map[string]any{
					"type":       "object",
					"properties": map[string]any{},
				},
			},
		},
	}

	// NOTE: claude-3-7-sonnet-20250219 is not support parallel tool calls with streaming
	// If you need to test parallel tool calls without streaming, you can use the following options:
	// * WithAnthropicBetaHeader("token-efficient-tools-2025-02-19") while initializing client
	// * llms.WithToolChoice(map[string]any{"type": "auto"}) while call GenerateContent
	llm := newHTTPRRClient(t, anthropic.WithModel("claude-sonnet-4-5"))

	content := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextContent{
					Text: "Calculate a 15% tip on a $50 bill and tell me who is the waiter in the format: '<waiterName> should get a tip of <tipAmount>'", //nolint:lll
				},
			},
		},
	}

	var (
		streamedContent      []string
		streamedDone         bool
		toolCallDetected     = make(map[string]struct{})
		respToolCallDetected = make(map[string]struct{})
		toolCalls            = make(map[string]*streaming.ToolCall)
	)

	streamingFunc := func(_ context.Context, chunk streaming.Chunk) error {
		switch chunk.Type {
		case streaming.ChunkTypeNone:
			t.Errorf("unexpected chunk type: %s", chunk.Type)
		case streaming.ChunkTypeText:
			streamedContent = append(streamedContent, chunk.Content)
		case streaming.ChunkTypeReasoning:
			// skip reasoning chunks
		case streaming.ChunkTypeToolCall:
			toolCall := chunk.ToolCall
			if toolCall.Name != "calculateTip" && toolCall.Name != "getWaiterName" {
				return fmt.Errorf("unexpected tool call: %s", toolCall.Name)
			}

			// Should not be more chunks after tool call is detected
			if _, ok := toolCallDetected[toolCall.Name]; ok {
				return fmt.Errorf("got extra chunk after tool call is detected")
			}

			if resToolCall, ok := toolCalls[toolCall.Name]; !ok {
				toolCalls[toolCall.Name] = &toolCall
			} else {
				streaming.AppendToolCall(toolCall, resToolCall)
			}
		case streaming.ChunkTypeDone:
			streamedDone = true
		}

		return nil
	}

	resp, err := llm.GenerateContent(
		t.Context(),
		content,
		llms.WithStreamingFunc(streamingFunc),
		llms.WithTools(availableTools),
		llms.WithTemperature(0.5),
	)
	require.NoError(t, err)

	// Check if tool call is complete
	for _, toolCall := range toolCalls {
		if _, err := toolCall.Parse(); err == nil {
			toolCallDetected[toolCall.Name] = struct{}{}
		}
	}

	// Verify the tool call result
	assert.Len(t, toolCallDetected, 2, "tool call not detected in streaming")
	toolCallResult, ok := toolCalls["calculateTip"]
	require.True(t, ok)
	calculateTipArgs, err := toolCallResult.Parse()
	require.NoError(t, err)
	assert.Equal(t, float64(15), calculateTipArgs["tipPercentage"])
	assert.Equal(t, float64(50), calculateTipArgs["billAmount"])

	// Verify the streamed content
	assert.True(t, streamedDone)

	fullContent := strings.Join(streamedContent, "")

	assert.NotEmpty(t, resp.Choices)
	for _, choice := range resp.Choices {
		if len(choice.Content) != 0 {
			assert.Equal(t, fullContent, choice.Content)
		}

		for _, toolCall := range choice.ToolCalls {
			streamedToolCall, ok := toolCalls[toolCall.FunctionCall.Name]
			assert.True(t, ok)
			assert.NotNil(t, streamedToolCall)
			assert.NotNil(t, toolCall.FunctionCall)
			assert.Equal(t, toolCall.ID, streamedToolCall.ID)
			assert.Equal(t, toolCall.FunctionCall.Name, streamedToolCall.Name)

			// Special case for tool call without arguments (there is empty string arguments)
			if toolCall.FunctionCall.Arguments != "" && streamedToolCall.Arguments != "" {
				assert.JSONEq(t, toolCall.FunctionCall.Arguments, streamedToolCall.Arguments)
			}

			respToolCallDetected[toolCall.FunctionCall.Name] = struct{}{}
		}
	}

	assert.Len(t, respToolCallDetected, 2, "tool call not detected in response")
}
