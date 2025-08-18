package bedrock_test

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/internal/httprr"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/bedrock"
)

func TestBedrockAnthropicToolCalling(t *testing.T) {
	ctx := context.Background()

	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "AWS_ACCESS_KEY_ID")

	rr := httprr.OpenForTest(t, http.DefaultTransport)
	defer rr.Close()

	// Only run tests in parallel when not recording (to avoid rate limits)
	if !rr.Recording() {
		t.Parallel()
	}

	// Configure AWS client to use httprr transport
	client, err := setUpTestWithTransport(rr)
	if err != nil {
		t.Fatal(err)
	}

	llm, err := bedrock.New(bedrock.WithClient(client))
	if err != nil {
		t.Fatal(err)
	}

	// Define a weather tool
	weatherTool := llms.Tool{
		Type: "function",
		Function: &llms.FunctionDefinition{
			Name:        "get_weather",
			Description: "Get the current weather for a location",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"location": map[string]interface{}{
						"type":        "string",
						"description": "The city and state, e.g. San Francisco, CA",
					},
					"unit": map[string]interface{}{
						"type":        "string",
						"enum":        []string{"celsius", "fahrenheit"},
						"description": "The unit of temperature",
					},
				},
				"required": []string{"location"},
			},
		},
	}

	t.Run("Anthropic Claude 3 with tool calling", func(t *testing.T) {
		// Skip if not recording and no credentials
		if !rr.Recording() && !hasAWSCredentials() {
			t.Skip("Skipping test: no AWS credentials and not recording")
		}

		msgs := []llms.MessageContent{
			llms.TextParts(llms.ChatMessageTypeHuman, "What's the weather like in New York?"),
		}

		resp, err := llm.GenerateContent(ctx, msgs,
			llms.WithModel(bedrock.ModelAnthropicClaudeV3Haiku),
			llms.WithTools([]llms.Tool{weatherTool}),
			llms.WithMaxTokens(512),
		)
		if err != nil {
			t.Fatal(err)
		}

		require.NotNil(t, resp)
		require.NotEmpty(t, resp.Choices)

		// Check if the model wants to use the tool
		choice := resp.Choices[0]
		if len(choice.ToolCalls) > 0 {
			t.Logf("Model requested tool call: %+v", choice.ToolCalls[0])
			
			// Verify tool call structure
			toolCall := choice.ToolCalls[0]
			require.Equal(t, "function", toolCall.Type)
			require.NotNil(t, toolCall.FunctionCall)
			require.Equal(t, "get_weather", toolCall.FunctionCall.Name)
			
			// Parse arguments
			var args map[string]interface{}
			err := json.Unmarshal([]byte(toolCall.FunctionCall.Arguments), &args)
			require.NoError(t, err)
			require.Contains(t, args, "location")
			
			// Simulate tool response
			toolResponse := llms.MessageContent{
				Role: llms.ChatMessageTypeTool,
				Parts: []llms.ContentPart{
					llms.ToolCallResponse{
						ToolCallID: toolCall.ID,
						Name:       "get_weather",
						Content:    "It's currently 72°F and sunny in New York",
					},
				},
			}
			
			// Continue conversation with tool response
			msgs = append(msgs, toolResponse)
			
			resp2, err := llm.GenerateContent(ctx, msgs,
				llms.WithModel(bedrock.ModelAnthropicClaudeV3Haiku),
				llms.WithMaxTokens(512),
			)
			require.NoError(t, err)
			require.NotNil(t, resp2)
			require.NotEmpty(t, resp2.Choices)
			
			t.Logf("Final response: %s", resp2.Choices[0].Content)
		} else {
			t.Logf("Model response without tool call: %s", choice.Content)
		}
	})

	t.Run("Multiple tools", func(t *testing.T) {
		// Skip if not recording and no credentials
		if !rr.Recording() && !hasAWSCredentials() {
			t.Skip("Skipping test: no AWS credentials and not recording")
		}

		calculatorTool := llms.Tool{
			Type: "function",
			Function: &llms.FunctionDefinition{
				Name:        "calculate",
				Description: "Perform mathematical calculations",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"expression": map[string]interface{}{
							"type":        "string",
							"description": "The mathematical expression to evaluate",
						},
					},
					"required": []string{"expression"},
				},
			},
		}

		msgs := []llms.MessageContent{
			llms.TextParts(llms.ChatMessageTypeHuman, "What's 123 * 456 and what's the weather in Paris?"),
		}

		resp, err := llm.GenerateContent(ctx, msgs,
			llms.WithModel(bedrock.ModelAnthropicClaudeV3Haiku),
			llms.WithTools([]llms.Tool{weatherTool, calculatorTool}),
			llms.WithMaxTokens(512),
		)
		if err != nil {
			t.Fatal(err)
		}

		require.NotNil(t, resp)
		require.NotEmpty(t, resp.Choices)
		
		// Check for multiple tool calls
		choice := resp.Choices[0]
		t.Logf("Number of tool calls: %d", len(choice.ToolCalls))
		for i, tc := range choice.ToolCalls {
			t.Logf("Tool call %d: %s(%s)", i, tc.FunctionCall.Name, tc.FunctionCall.Arguments)
		}
	})
}

func hasAWSCredentials() bool {
	// Check if AWS credentials are available
	_, hasKey := os.LookupEnv("AWS_ACCESS_KEY_ID")
	_, hasProfile := os.LookupEnv("AWS_PROFILE")
	return hasKey || hasProfile
}