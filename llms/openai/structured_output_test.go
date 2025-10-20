package openai

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai/internal/openaiclient"
)

func TestStructuredOutputObjectSchema(t *testing.T) {
	ctx := context.Background()
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

	rsp, err := llm.GenerateContent(ctx, content)
	require.NoError(t, err)

	assert.NotEmpty(t, rsp.Choices)
	c1 := rsp.Choices[0]
	assert.Regexp(t, "\"final_answer\":", strings.ToLower(c1.Content))
}

func TestStructuredOutputObjectAndArraySchema(t *testing.T) {
	ctx := context.Background()
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

	rsp, err := llm.GenerateContent(ctx, content)
	require.NoError(t, err)

	assert.NotEmpty(t, rsp.Choices)
	c1 := rsp.Choices[0]
	assert.Regexp(t, "\"steps\":", strings.ToLower(c1.Content))
}

func TestStructuredOutputFunctionCalling(t *testing.T) {
	ctx := context.Background()
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
		ctx,
		content,
		llms.WithTools(toolList),
	)
	require.NoError(t, err)

	assert.NotEmpty(t, rsp.Choices)
	c1 := rsp.Choices[0]
	assert.Regexp(t, "\"search_engine\":", c1.ToolCalls[0].FunctionCall.Arguments)
	assert.Regexp(t, "\"search_query\":", c1.ToolCalls[0].FunctionCall.Arguments)
}

// Test schema conversion - simple object
func TestConvertStructuredOutputSimpleObject(t *testing.T) {
	schema := &llms.StructuredOutputDefinition{
		Name:        "user_info",
		Description: "Extract user information",
		Schema: &llms.StructuredOutputSchema{
			Type: llms.SchemaTypeObject,
			Properties: map[string]*llms.StructuredOutputSchema{
				"name": {
					Type:        llms.SchemaTypeString,
					Description: "User's full name",
				},
				"age": {
					Type:        llms.SchemaTypeInteger,
					Description: "User's age in years",
				},
			},
			Required:             []string{"name", "age"},
			AdditionalProperties: false,
		},
		Strict: true,
	}

	responseFormat, err := convertStructuredOutputToResponseFormat(schema)
	require.NoError(t, err)
	require.NotNil(t, responseFormat)

	assert.Equal(t, "json_schema", responseFormat.Type)
	assert.NotNil(t, responseFormat.JSONSchema)
	assert.Equal(t, "user_info", responseFormat.JSONSchema.Name)
	assert.True(t, responseFormat.JSONSchema.Strict)
	assert.NotNil(t, responseFormat.JSONSchema.Schema)
	assert.Equal(t, "object", responseFormat.JSONSchema.Schema.Type)
	assert.Len(t, responseFormat.JSONSchema.Schema.Properties, 2)
	assert.Contains(t, responseFormat.JSONSchema.Schema.Required, "name")
	assert.Contains(t, responseFormat.JSONSchema.Schema.Required, "age")
	assert.False(t, responseFormat.JSONSchema.Schema.AdditionalProperties)
}

// Test schema conversion - nested objects
func TestConvertStructuredOutputNestedObject(t *testing.T) {
	schema := &llms.StructuredOutputDefinition{
		Name: "company_info",
		Schema: &llms.StructuredOutputSchema{
			Type: llms.SchemaTypeObject,
			Properties: map[string]*llms.StructuredOutputSchema{
				"name": {
					Type: llms.SchemaTypeString,
				},
				"address": {
					Type: llms.SchemaTypeObject,
					Properties: map[string]*llms.StructuredOutputSchema{
						"street": {Type: llms.SchemaTypeString},
						"city":   {Type: llms.SchemaTypeString},
						"zip":    {Type: llms.SchemaTypeString},
					},
					Required: []string{"city"},
				},
			},
			Required: []string{"name"},
		},
		Strict: false,
	}

	responseFormat, err := convertStructuredOutputToResponseFormat(schema)
	require.NoError(t, err)
	require.NotNil(t, responseFormat)

	assert.Equal(t, "json_schema", responseFormat.Type)
	assert.False(t, responseFormat.JSONSchema.Strict)
	assert.NotNil(t, responseFormat.JSONSchema.Schema.Properties["address"])
	assert.Equal(t, "object", responseFormat.JSONSchema.Schema.Properties["address"].Type)
	assert.Len(t, responseFormat.JSONSchema.Schema.Properties["address"].Properties, 3)
	assert.Contains(t, responseFormat.JSONSchema.Schema.Properties["address"].Required, "city")
}

// Test schema conversion - arrays
func TestConvertStructuredOutputArray(t *testing.T) {
	schema := &llms.StructuredOutputDefinition{
		Name: "tags_list",
		Schema: &llms.StructuredOutputSchema{
			Type: llms.SchemaTypeObject,
			Properties: map[string]*llms.StructuredOutputSchema{
				"tags": {
					Type: llms.SchemaTypeArray,
					Items: &llms.StructuredOutputSchema{
						Type: llms.SchemaTypeString,
					},
				},
			},
			Required: []string{"tags"},
		},
	}

	responseFormat, err := convertStructuredOutputToResponseFormat(schema)
	require.NoError(t, err)
	require.NotNil(t, responseFormat)

	tagsProperty := responseFormat.JSONSchema.Schema.Properties["tags"]
	assert.NotNil(t, tagsProperty)
	assert.Equal(t, "array", tagsProperty.Type)
	assert.NotNil(t, tagsProperty.Items)
	assert.Equal(t, "string", tagsProperty.Items.Type)
}

// Test schema conversion - enums
func TestConvertStructuredOutputEnum(t *testing.T) {
	schema := &llms.StructuredOutputDefinition{
		Name: "choice",
		Schema: &llms.StructuredOutputSchema{
			Type: llms.SchemaTypeObject,
			Properties: map[string]*llms.StructuredOutputSchema{
				"color": {
					Type: llms.SchemaTypeString,
					Enum: []string{"red", "green", "blue"},
				},
			},
		},
	}

	responseFormat, err := convertStructuredOutputToResponseFormat(schema)
	require.NoError(t, err)
	require.NotNil(t, responseFormat)

	colorProperty := responseFormat.JSONSchema.Schema.Properties["color"]
	assert.NotNil(t, colorProperty)
	assert.Len(t, colorProperty.Enum, 3)
	assert.Contains(t, colorProperty.Enum, "red")
	assert.Contains(t, colorProperty.Enum, "green")
	assert.Contains(t, colorProperty.Enum, "blue")
}

// Test unified WithStructuredOutput API
func TestUnifiedStructuredOutputAPI(t *testing.T) {
	t.Skip("Skipping integration test - run with OPENAI_API_KEY to test")
	ctx := context.Background()

	// Define schema using unified llms API
	schema := &llms.StructuredOutputDefinition{
		Name:        "math_solution",
		Description: "Solution to a math problem",
		Schema: &llms.StructuredOutputSchema{
			Type: llms.SchemaTypeObject,
			Properties: map[string]*llms.StructuredOutputSchema{
				"steps": {
					Type:        llms.SchemaTypeArray,
					Description: "Step-by-step solution",
					Items: &llms.StructuredOutputSchema{
						Type: llms.SchemaTypeString,
					},
				},
				"final_answer": {
					Type:        llms.SchemaTypeString,
					Description: "The final answer",
				},
			},
			Required:             []string{"final_answer", "steps"},
			AdditionalProperties: false,
		},
		Strict: true,
	}

	llm := newTestClient(t, WithModel("gpt-4o-2024-08-06"))

	content := []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeSystem,
			Parts: []llms.ContentPart{llms.TextContent{Text: "You are a math tutor."}},
		},
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextContent{Text: "Solve 15 + 27"}},
		},
	}

	// Use unified WithStructuredOutput option
	rsp, err := llm.GenerateContent(ctx, content,
		llms.WithStructuredOutput(schema),
	)
	require.NoError(t, err)
	require.NotEmpty(t, rsp.Choices)

	c1 := rsp.Choices[0]
	// Verify response contains expected fields
	assert.Regexp(t, "\"final_answer\":", strings.ToLower(c1.Content))
	assert.Regexp(t, "\"steps\":", strings.ToLower(c1.Content))

	// Verify we can parse it as valid JSON
	var result map[string]interface{}
	err = json.Unmarshal([]byte(c1.Content), &result)
	require.NoError(t, err)

	// Verify structure
	assert.Contains(t, result, "final_answer")
	assert.Contains(t, result, "steps")
	assert.IsType(t, []interface{}{}, result["steps"])
}
