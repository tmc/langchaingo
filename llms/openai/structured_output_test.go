package openai

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vendasta/langchaingo/llms"
	"github.com/vendasta/langchaingo/llms/openai/internal/openaiclient"
)

func TestStructuredOutputObjectSchema(t *testing.T) {
	t.Parallel()
	responseFormat := &ResponseFormat{
		Type: "json_schema",
		JSONSchema: &ResponseFormatJSONSchema{
			Name:   "math_schema",
			Strict: true,
			Schema: &ResponseFormatJSONSchemaProperty{
				Type: "object",
				Properties: map[string]*ResponseFormatJSONSchemaProperty{
					"final_answer": {
						Type: "string",
					},
				},
				AdditionalProperties: false,
				Required:             []string{"final_answer"},
			},
		},
	}
	llm := newTestClient(
		t,
		WithModel("gpt-4o-2024-08-06"),
		WithResponseFormat(responseFormat),
	)

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

	rsp, err := llm.GenerateContent(context.Background(), content)
	require.NoError(t, err)

	assert.NotEmpty(t, rsp.Choices)
	c1 := rsp.Choices[0]
	assert.Regexp(t, "\"final_answer\":", strings.ToLower(c1.Content))
}

func TestStructuredOutputObjectAndArraySchema(t *testing.T) {
	t.Parallel()
	responseFormat := &ResponseFormat{
		Type: "json_schema",
		JSONSchema: &ResponseFormatJSONSchema{
			Name:   "math_schema",
			Strict: true,
			Schema: &ResponseFormatJSONSchemaProperty{
				Type: "object",
				Properties: map[string]*ResponseFormatJSONSchemaProperty{
					"steps": {
						Type: "array",
						Items: &ResponseFormatJSONSchemaProperty{
							Type: "string",
						},
					},
					"final_answer": {
						Type: "string",
					},
				},
				AdditionalProperties: false,
				Required:             []string{"final_answer", "steps"},
			},
		},
	}
	llm := newTestClient(
		t,
		WithModel("gpt-4o-2024-08-06"),
		WithResponseFormat(responseFormat),
	)

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

	rsp, err := llm.GenerateContent(context.Background(), content)
	require.NoError(t, err)

	assert.NotEmpty(t, rsp.Choices)
	c1 := rsp.Choices[0]
	assert.Regexp(t, "\"steps\":", strings.ToLower(c1.Content))
}

func TestStructuredOutputFunctionCalling(t *testing.T) {
	t.Parallel()
	llm := newTestClient(
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
			Role:  llms.ChatMessageTypeGeneric,
			Parts: []llms.ContentPart{llms.TextContent{Text: "What is the age of Bob Odenkirk, a famous comedy screenwriter and an actor."}},
		},
	}

	rsp, err := llm.GenerateContent(
		context.Background(),
		content,
		llms.WithTools(toolList),
	)
	require.NoError(t, err)

	assert.NotEmpty(t, rsp.Choices)
	c1 := rsp.Choices[0]
	assert.Regexp(t, "\"search_engine\":", c1.ToolCalls[0].FunctionCall.Arguments)
	assert.Regexp(t, "\"search_query\":", c1.ToolCalls[0].FunctionCall.Arguments)
}
