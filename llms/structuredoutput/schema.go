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
	t, _ := m["type"].(string)
	props, hasProps := m["properties"].(map[string]any)
	if t == "object" || hasProps {
		if ap, ok := m["additionalProperties"].(bool); !ok || ap {
			return fmt.Errorf("%w: every object schema must explicitly set additionalProperties:false",
				llms.ErrStructuredOutputConfig)
		}
	}
	for _, v := range props {
		if err := requireClosed(v); err != nil {
			return err
		}
	}
	for _, key := range []string{"$defs", "definitions"} {
		if defs, ok := m[key].(map[string]any); ok {
			for _, v := range defs {
				if err := requireClosed(v); err != nil {
					return err
				}
			}
		}
	}
	if items, ok := m["items"]; ok {
		if err := requireClosed(items); err != nil {
			return err
		}
	}
	for _, key := range []string{"anyOf", "oneOf", "allOf"} {
		if arr, ok := m[key].([]any); ok {
			for _, v := range arr {
				if err := requireClosed(v); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
