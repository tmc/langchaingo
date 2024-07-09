package anthropic

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/llms"
)

func newTestClient(t *testing.T, opts ...Option) *LLM {
	t.Helper()
	var apiKey string

	if apiKey = os.Getenv("ANTHROPIC_API_KEY"); apiKey == "" {
		t.Skip("ANTHROPIC_API_KEY not set")
		return nil
	}

	llm, err := New(opts...)
	require.NoError(t, err)

	return llm
}

func TestGenerateContent(t *testing.T) {
	t.Parallel()
	llm := newTestClient(t)

	parts := []llms.ContentPart{
		llms.TextContent{Text: "How many feet are in a nautical mile?"},
	}
	content := []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: parts,
		},
	}

	rsp, err := llm.GenerateContent(context.Background(), content)
	require.NoError(t, err)

	assert.NotEmpty(t, rsp.Choices)
	c1 := rsp.Choices[0]
	assert.Regexp(t, "feet", strings.ToLower(c1.Content))
}

func TestGenerateContentWithTool(t *testing.T) {
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

	llm := newTestClient(t, WithModel("claude-3-haiku-20240307"))

	contents := []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextContent{Text: "What is the weather in Boston?"}},
		},
	}

	// Ask a questions about the weather and let the model know that the tool is available
	rsp, err := llm.GenerateContent(context.Background(), contents, llms.WithTools(availableTools))
	require.NoError(t, err)

	// Expect a tool call in the response
	require.NotEmpty(t, rsp.Choices)
	choice := rsp.Choices[0]
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
	rsp, err = llm.GenerateContent(context.Background(), contents, llms.WithTools(availableTools))
	require.NoError(t, err)

	require.NotEmpty(t, rsp.Choices)
	choice = rsp.Choices[0]
	assert.Regexp(t, "72", choice.Content)
}
