package openai

import (
	"fmt"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai/internal/openaiclient"
)

// convertStructuredOutputToResponseFormat converts a unified StructuredOutputDefinition
// to OpenAI's ResponseFormat structure for json_schema mode.
func convertStructuredOutputToResponseFormat(def *llms.StructuredOutputDefinition) (*openaiclient.ResponseFormat, error) {
	if def == nil {
		return nil, fmt.Errorf("structured output definition cannot be nil")
	}

	if def.Schema == nil {
		return nil, fmt.Errorf("structured output schema cannot be nil")
	}

	// Validate the schema before conversion
	if err := llms.ValidateSchema(def.Schema); err != nil {
		return nil, fmt.Errorf("invalid schema: %w", err)
	}

	// Convert the unified schema to OpenAI's schema format
	openaiSchema, err := convertSchemaToOpenAI(def.Schema)
	if err != nil {
		return nil, fmt.Errorf("failed to convert schema: %w", err)
	}

	// Build the ResponseFormat
	return &openaiclient.ResponseFormat{
		Type: "json_schema",
		JSONSchema: &openaiclient.ResponseFormatJSONSchema{
			Name:   def.Name,
			Strict: def.Strict,
			Schema: openaiSchema,
		},
	}, nil
}

// convertSchemaToOpenAI converts a unified StructuredOutputSchema to OpenAI's
// ResponseFormatJSONSchemaProperty structure.
func convertSchemaToOpenAI(schema *llms.StructuredOutputSchema) (*openaiclient.ResponseFormatJSONSchemaProperty, error) {
	if schema == nil {
		return nil, fmt.Errorf("schema cannot be nil")
	}

	prop := &openaiclient.ResponseFormatJSONSchemaProperty{
		Type:                 string(schema.Type),
		Description:          schema.Description,
		AdditionalProperties: schema.AdditionalProperties,
	}

	// Handle enum values
	if len(schema.Enum) > 0 {
		prop.Enum = make([]interface{}, len(schema.Enum))
		for i, v := range schema.Enum {
			prop.Enum[i] = v
		}
	}

	// Handle object properties
	if schema.Type == llms.SchemaTypeObject && len(schema.Properties) > 0 {
		prop.Properties = make(map[string]*openaiclient.ResponseFormatJSONSchemaProperty)
		for name, subSchema := range schema.Properties {
			convertedProp, err := convertSchemaToOpenAI(subSchema)
			if err != nil {
				return nil, fmt.Errorf("failed to convert property %q: %w", name, err)
			}
			prop.Properties[name] = convertedProp
		}

		// Set required fields
		if len(schema.Required) > 0 {
			prop.Required = schema.Required
		}
	}

	// Handle array items
	if schema.Type == llms.SchemaTypeArray && schema.Items != nil {
		convertedItems, err := convertSchemaToOpenAI(schema.Items)
		if err != nil {
			return nil, fmt.Errorf("failed to convert array items: %w", err)
		}
		prop.Items = convertedItems
	}

	return prop, nil
}
