package openai

import (
	"encoding/json"
	"testing"

	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/llms/openai/internal/openaiclient"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStructuredOutputObjectSchema(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	llm := newTestOpenAIClient(t, WithModel("gpt-4o-2024-08-06"))

	schema := json.RawMessage(`{
		"type": "object",
		"properties": {"final_answer": {"type": "string"}},
		"required": ["final_answer"],
		"additionalProperties": false
	}`)

	content := []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeSystem,
			Parts: []llms.ContentPart{llms.TextContent{Text: "You are a student taking a math exam."}},
		},
		{
			Role:  llms.ChatMessageTypeGeneric,
			Parts: []llms.ContentPart{llms.TextContent{Text: "Solve 2 + 2"}},
		},
	}

	resp, err := llm.GenerateContent(ctx, content,
		llms.WithStructuredOutput(llms.StructuredOutputConfig{Name: "math_schema", Schema: schema}))
	require.NoError(t, err)

	require.NotEmpty(t, resp.Choices)
	// The general path guarantees the content parses and matches the schema.
	var out struct {
		FinalAnswer string `json:"final_answer"`
	}
	require.NoError(t, json.Unmarshal([]byte(resp.Choices[0].Content), &out))
	assert.NotEmpty(t, out.FinalAnswer)
}

func TestStructuredOutputObjectAndArraySchema(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	llm := newTestOpenAIClient(t, WithModel("gpt-4o-2024-08-06"))

	schema := json.RawMessage(`{
		"type": "object",
		"properties": {
			"steps": {"type": "array", "items": {"type": "string"}},
			"final_answer": {"type": "string"}
		},
		"required": ["steps", "final_answer"],
		"additionalProperties": false
	}`)

	content := []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeSystem,
			Parts: []llms.ContentPart{llms.TextContent{Text: "You are a student taking a math exam."}},
		},
		{
			Role:  llms.ChatMessageTypeGeneric,
			Parts: []llms.ContentPart{llms.TextContent{Text: "Solve 2 + 2"}},
		},
	}

	resp, err := llm.GenerateContent(ctx, content,
		llms.WithStructuredOutput(llms.StructuredOutputConfig{Name: "math_schema", Schema: schema}))
	require.NoError(t, err)

	require.NotEmpty(t, resp.Choices)
	var out struct {
		Steps       []string `json:"steps"`
		FinalAnswer string   `json:"final_answer"`
	}
	require.NoError(t, json.Unmarshal([]byte(resp.Choices[0].Content), &out))
	assert.NotEmpty(t, out.FinalAnswer)
	assert.NotEmpty(t, out.Steps)
}

func TestStructuredOutputFunctionCalling(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	llm := newTestOpenAIClient(
		t,
		WithModel("gpt-4o-2024-08-06"),
	)

	toolList := []llms.Tool{
		{
			Type: string(openaiclient.ToolTypeFunction),
			Function: &llms.FunctionDefinition{
				Name:        "search",
				Description: "Search by the web search engine",
				Parameters: json.RawMessage(
					`{
					"type": "object",
					"properties" : {
						"search_engine" : {
							"type" : "string",
							"enum" : ["google", "duckduckgo", "bing"]
						},
						"search_query" : {
							"type" : "string"
						}
					},
					"required":["search_engine", "search_query"],
					"additionalProperties": false
				}`),
				Strict: true,
			},
		},
	}

	content := []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeSystem,
			Parts: []llms.ContentPart{llms.TextContent{Text: "You are a helpful assistant"}},
		},
		{
			Role: llms.ChatMessageTypeGeneric,
			Parts: []llms.ContentPart{
				llms.TextContent{
					Text: "What is the age of Bob Odenkirk, a famous comedy screenwriter and an actor.",
				},
			},
		},
	}

	resp, err := llm.GenerateContent(
		ctx,
		content,
		llms.WithTools(toolList),
	)
	require.NoError(t, err)

	assert.NotEmpty(t, resp.Choices)
	c1 := resp.Choices[0]
	assert.Regexp(t, "\"search_engine\":", c1.ToolCalls[0].FunctionCall.Arguments)
	assert.Regexp(t, "\"search_query\":", c1.ToolCalls[0].FunctionCall.Arguments)
}
