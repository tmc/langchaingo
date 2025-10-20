package anthropic

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/llms"
)

func TestSchemaToTool(t *testing.T) {
	tests := []struct {
		name     string
		schema   *llms.StructuredOutputDefinition
		expected map[string]any
	}{
		{
			name: "simple object schema",
			schema: &llms.StructuredOutputDefinition{
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
					Required: []string{"name", "age"},
				},
			},
			expected: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"name": map[string]any{
						"type":        "string",
						"description": "User's full name",
					},
					"age": map[string]any{
						"type":        "integer",
						"description": "User's age in years",
					},
				},
				"required":             []string{"name", "age"},
				"additionalProperties": false,
			},
		},
		{
			name: "array schema",
			schema: &llms.StructuredOutputDefinition{
				Name:        "recipe",
				Description: "A cooking recipe",
				Schema: &llms.StructuredOutputSchema{
					Type: llms.SchemaTypeObject,
					Properties: map[string]*llms.StructuredOutputSchema{
						"title": {
							Type: llms.SchemaTypeString,
						},
						"ingredients": {
							Type: llms.SchemaTypeArray,
							Items: &llms.StructuredOutputSchema{
								Type: llms.SchemaTypeString,
							},
						},
					},
					Required: []string{"title", "ingredients"},
				},
			},
			expected: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"title": map[string]any{
						"type": "string",
					},
					"ingredients": map[string]any{
						"type": "array",
						"items": map[string]any{
							"type": "string",
						},
					},
				},
				"required":             []string{"title", "ingredients"},
				"additionalProperties": false,
			},
		},
		{
			name: "enum schema",
			schema: &llms.StructuredOutputDefinition{
				Name: "status",
				Schema: &llms.StructuredOutputSchema{
					Type: llms.SchemaTypeObject,
					Properties: map[string]*llms.StructuredOutputSchema{
						"status": {
							Type: llms.SchemaTypeString,
							Enum: []string{"pending", "active", "completed"},
						},
					},
				},
			},
			expected: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"status": map[string]any{
						"type": "string",
						"enum": []string{"pending", "active", "completed"},
					},
				},
				"additionalProperties": false,
			},
		},
		{
			name: "number constraints",
			schema: &llms.StructuredOutputDefinition{
				Name: "rating",
				Schema: &llms.StructuredOutputSchema{
					Type: llms.SchemaTypeObject,
					Properties: map[string]*llms.StructuredOutputSchema{
						"score": {
							Type:    llms.SchemaTypeNumber,
							Minimum: func() *float64 { v := 0.0; return &v }(),
							Maximum: func() *float64 { v := 10.0; return &v }(),
						},
					},
				},
			},
			expected: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"score": map[string]any{
						"type":    "number",
						"minimum": 0.0,
						"maximum": 10.0,
					},
				},
				"additionalProperties": false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tool := schemaToTool(tt.schema)

			assert.Equal(t, tt.schema.Name, tool.Name)
			assert.Equal(t, tt.schema.Description, tool.Description)

			// Compare the input schema
			inputSchema, ok := tool.InputSchema.(map[string]any)
			require.True(t, ok, "InputSchema should be map[string]any")

			assert.Equal(t, tt.expected, inputSchema)
		})
	}
}

func TestConvertSchemaToJSONSchema(t *testing.T) {
	tests := []struct {
		name     string
		schema   *llms.StructuredOutputSchema
		expected map[string]any
	}{
		{
			name:     "nil schema",
			schema:   nil,
			expected: nil,
		},
		{
			name: "simple string",
			schema: &llms.StructuredOutputSchema{
				Type:        llms.SchemaTypeString,
				Description: "A string field",
			},
			expected: map[string]any{
				"type":        "string",
				"description": "A string field",
			},
		},
		{
			name: "string with constraints",
			schema: &llms.StructuredOutputSchema{
				Type:      llms.SchemaTypeString,
				MinLength: func() *int { v := 5; return &v }(),
				MaxLength: func() *int { v := 100; return &v }(),
				Pattern:   "^[A-Z]",
			},
			expected: map[string]any{
				"type":      "string",
				"minLength": 5,
				"maxLength": 100,
				"pattern":   "^[A-Z]",
			},
		},
		{
			name: "array with constraints",
			schema: &llms.StructuredOutputSchema{
				Type: llms.SchemaTypeArray,
				Items: &llms.StructuredOutputSchema{
					Type: llms.SchemaTypeInteger,
				},
				MinItems: func() *int { v := 1; return &v }(),
				MaxItems: func() *int { v := 10; return &v }(),
			},
			expected: map[string]any{
				"type": "array",
				"items": map[string]any{
					"type": "integer",
				},
				"minItems": 1,
				"maxItems": 10,
			},
		},
		{
			name: "nested object",
			schema: &llms.StructuredOutputSchema{
				Type: llms.SchemaTypeObject,
				Properties: map[string]*llms.StructuredOutputSchema{
					"user": {
						Type: llms.SchemaTypeObject,
						Properties: map[string]*llms.StructuredOutputSchema{
							"id": {
								Type: llms.SchemaTypeInteger,
							},
							"name": {
								Type: llms.SchemaTypeString,
							},
						},
						Required: []string{"id"},
					},
				},
			},
			expected: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"user": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"id": map[string]any{
								"type": "integer",
							},
							"name": map[string]any{
								"type": "string",
							},
						},
						"required":             []string{"id"},
						"additionalProperties": false,
					},
				},
				"additionalProperties": false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertSchemaToJSONSchema(tt.schema)
			assert.Equal(t, tt.expected, result)
		})
	}
}
