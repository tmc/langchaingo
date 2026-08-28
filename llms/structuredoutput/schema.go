package structuredoutput

import (
	"encoding/json"
	"fmt"

	"github.com/vxcontrol/langchaingo/llms"
)

// RequireClosedObjects verifies that every object node in the schema explicitly
// sets additionalProperties:false. Anthropic and Amazon Bedrock structured outputs
// reject an object schema that omits it (a ValidationException 400), and this is
// reliably detectable locally, so it is turned into a typed configuration error
// before the request is sent instead of an opaque provider 4xx. It wraps
// llms.ErrStructuredOutputConfig and does not otherwise constrain the schema.
func RequireClosedObjects(schema json.RawMessage) error {
	var node any
	if err := json.Unmarshal(schema, &node); err != nil {
		return fmt.Errorf("%w: %w", llms.ErrStructuredOutputConfig, err)
	}
	return requireClosed(node)
}

func requireClosed(node any) error {
	m, ok := node.(map[string]any)
	if !ok {
		return nil
	}
	// A node is an object schema when its type is "object" or it declares
	// properties. Such a node must close itself with additionalProperties:false.
	_, hasProps := m["properties"].(map[string]any)
	if declaresObject(m["type"]) || hasProps {
		if ap, ok := m["additionalProperties"].(bool); !ok || ap {
			return fmt.Errorf("%w: every object schema must explicitly set additionalProperties:false",
				llms.ErrStructuredOutputConfig)
		}
	}
	for _, key := range []string{"properties", "patternProperties", "$defs", "definitions", "dependentSchemas"} {
		if group, ok := m[key].(map[string]any); ok {
			for _, v := range group {
				if err := requireClosedValue(v); err != nil {
					return err
				}
			}
		}
	}
	for _, key := range []string{
		"items", "prefixItems", "additionalProperties", "anyOf", "oneOf", "allOf",
		"not", "if", "then", "else", "contains", "propertyNames",
		"unevaluatedItems", "unevaluatedProperties",
	} {
		if err := requireClosedValue(m[key]); err != nil {
			return err
		}
	}
	return nil
}

func requireClosedValue(v any) error {
	switch node := v.(type) {
	case map[string]any:
		return requireClosed(node)
	case []any:
		for _, item := range node {
			if err := requireClosedValue(item); err != nil {
				return err
			}
		}
	}
	return nil
}

func declaresObject(v any) bool {
	switch t := v.(type) {
	case string:
		return t == "object"
	case []any:
		for _, item := range t {
			if name, ok := item.(string); ok && name == "object" {
				return true
			}
		}
	}
	return false
}
