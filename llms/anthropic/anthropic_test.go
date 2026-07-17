package anthropic_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/llms/anthropic"
	"github.com/vxcontrol/langchaingo/llms/streaming"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func generateAgainst(t *testing.T, body string) (*llms.ContentResponse, error) {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)

	llm, err := anthropic.New(anthropic.WithToken("k"), anthropic.WithBaseURL(srv.URL), anthropic.WithModel("claude-fable-5"))
	require.NoError(t, err)
	return llm.GenerateContent(t.Context(),
		[]llms.MessageContent{{Role: llms.ChatMessageTypeHuman, Parts: []llms.ContentPart{llms.TextPart("hi")}}})
}

// A refusal (stop_reason "refusal") must surface as a typed ErrModelRefusal
// carrying the refusal text, distinct from an empty/failed response.
func TestAnthropic_Refusal(t *testing.T) {
	t.Parallel()

	t.Run("refusal with text", func(t *testing.T) {
		t.Parallel()
		_, err := generateAgainst(t, `{"id":"m","type":"message","role":"assistant","model":"claude-fable-5",`+
			`"content":[{"type":"text","text":"I can't continue this story."}],"stop_reason":"refusal","usage":{"input_tokens":1,"output_tokens":1}}`)
		var refusal *anthropic.ErrModelRefusal
		require.ErrorAs(t, err, &refusal)
		assert.Equal(t, "I can't continue this story.", refusal.Message)
	})

	t.Run("refusal carries stop_details and usage", func(t *testing.T) {
		t.Parallel()
		_, err := generateAgainst(t, `{"id":"m","type":"message","role":"assistant","model":"claude-fable-5",`+
			`"content":[],"stop_reason":"refusal",`+
			`"stop_details":{"category":"cyber","explanation":"declined cyber content"},`+
			`"usage":{"input_tokens":42,"output_tokens":7}}`)
		var refusal *anthropic.ErrModelRefusal
		require.ErrorAs(t, err, &refusal)
		assert.Equal(t, "cyber", refusal.Category)
		assert.Equal(t, "declined cyber content", refusal.Explanation)
		assert.Equal(t, 42, refusal.InputTokens, "billed input tokens must be preserved")
		assert.Equal(t, 7, refusal.OutputTokens, "billed output tokens must be preserved")
	})

	t.Run("empty refusal is a refusal, not an empty response", func(t *testing.T) {
		t.Parallel()
		_, err := generateAgainst(t, `{"id":"m","type":"message","role":"assistant","model":"claude-fable-5",`+
			`"content":[],"stop_reason":"refusal","usage":{"input_tokens":1,"output_tokens":1}}`)
		var refusal *anthropic.ErrModelRefusal
		require.ErrorAs(t, err, &refusal, "empty-content refusal must be ErrModelRefusal, not ErrEmptyResponse")
		assert.False(t, errors.Is(err, anthropic.ErrEmptyResponse))
	})

	t.Run("normal end_turn is unaffected", func(t *testing.T) {
		t.Parallel()
		resp, err := generateAgainst(t, `{"id":"m","type":"message","role":"assistant","model":"claude-fable-5",`+
			`"content":[{"type":"text","text":"ok"}],"stop_reason":"end_turn","usage":{"input_tokens":1,"output_tokens":1}}`)
		require.NoError(t, err)
		assert.Equal(t, "ok", resp.Choices[0].Content)
	})
}

// TestAnthropic_StructuredOutputRealAPI exercises structured output against the
// real backend (httprr-recorded): a schema-constrained call must return JSON that
// matches the schema, and a schema-less WithJSONMode call must still work without
// applying any schema guarantee.
func TestAnthropic_StructuredOutputRealAPI(t *testing.T) {
	t.Parallel()

	t.Run("object schema is honored", func(t *testing.T) {
		t.Parallel()
		llm := newHTTPRRClient(t, anthropic.WithModel("claude-sonnet-4-5"))

		schema := json.RawMessage(`{
			"type": "object",
			"properties": {
				"country": {"type": "string"},
				"capital": {"type": "string"}
			},
			"required": ["country", "capital"],
			"additionalProperties": false
		}`)
		messages := []llms.MessageContent{{
			Role:  llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextPart("What is the capital of France?")},
		}}

		resp, err := llm.GenerateContent(t.Context(), messages,
			llms.WithStructuredOutput(llms.StructuredOutputConfig{Name: "capital", Schema: schema}),
			llms.WithMaxTokens(256),
		)
		require.NoError(t, err)
		require.NotEmpty(t, resp.Choices)

		var out struct {
			Country string `json:"country"`
			Capital string `json:"capital"`
		}
		require.NoError(t, json.Unmarshal([]byte(resp.Choices[0].Content), &out),
			"structured output must be valid JSON")
		assert.NotEmpty(t, out.Country)
		assert.Contains(t, strings.ToLower(out.Capital), "paris")
	})

	t.Run("array schema is honored", func(t *testing.T) {
		t.Parallel()
		llm := newHTTPRRClient(t, anthropic.WithModel("claude-sonnet-4-5"))

		schema := json.RawMessage(`{
			"type": "object",
			"properties": {
				"planets": {"type": "array", "items": {"type": "string"}}
			},
			"required": ["planets"],
			"additionalProperties": false
		}`)
		messages := []llms.MessageContent{{
			Role:  llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextPart("List the first three planets from the sun.")},
		}}

		resp, err := llm.GenerateContent(t.Context(), messages,
			llms.WithStructuredOutput(llms.StructuredOutputConfig{Name: "planets", Schema: schema}),
			llms.WithMaxTokens(256),
		)
		require.NoError(t, err)
		require.NotEmpty(t, resp.Choices)

		var out struct {
			Planets []string `json:"planets"`
		}
		require.NoError(t, json.Unmarshal([]byte(resp.Choices[0].Content), &out))
		assert.NotEmpty(t, out.Planets)
	})

	t.Run("schema-less JSONMode still works without a schema guarantee", func(t *testing.T) {
		t.Parallel()
		llm := newHTTPRRClient(t, anthropic.WithModel("claude-sonnet-4-5"))

		messages := []llms.MessageContent{{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextPart("Reply with only a JSON object with keys country and capital for France."),
			},
		}}

		// WithJSONMode alone must not error and must not apply structured output on
		// Anthropic (there is no json_object equivalent): a normal answer comes back.
		resp, err := llm.GenerateContent(t.Context(), messages,
			llms.WithJSONMode(),
			llms.WithMaxTokens(256),
		)
		require.NoError(t, err)
		require.NotEmpty(t, resp.Choices)
		assert.NotEmpty(t, resp.Choices[0].Content)
	})
}

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
