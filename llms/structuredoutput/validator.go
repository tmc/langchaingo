package structuredoutput

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v6"

	"github.com/vxcontrol/langchaingo/llms"
)

// Compiled is a schema compiled once so multiple choices of one response reuse it.
type Compiled struct {
	schema *jsonschema.Schema
}

// Compile compiles a raw JSON Schema document. Without $schema the document is
// treated as Draft 2020-12. The schema is used verbatim; no keywords are injected
// or stripped.
func Compile(schema json.RawMessage) (*Compiled, error) {
	doc, err := jsonschema.UnmarshalJSON(bytes.NewReader(schema))
	if err != nil {
		return nil, fmt.Errorf("parse schema: %w", err)
	}
	doc = admitNullable(doc)
	const resource = "structuredoutput:schema"
	comp := jsonschema.NewCompiler()
	comp.UseLoader(jsonschema.SchemeURLLoader{})
	if err := comp.AddResource(resource, doc); err != nil {
		return nil, fmt.Errorf("load schema: %w", err)
	}
	sch, err := comp.Compile(resource)
	if err != nil {
		return nil, fmt.Errorf("compile schema: %w", err)
	}
	return &Compiled{schema: sch}, nil
}

func admitNullable(node any) any {
	schema, ok := node.(map[string]any)
	if !ok {
		return node
	}
	for _, key := range schemaGroupKeywords {
		group, ok := schema[key].(map[string]any)
		if !ok {
			continue
		}
		for name, member := range group {
			group[name] = admitNullable(member)
		}
	}
	for _, key := range schemaValueKeywords {
		if value, ok := schema[key]; ok {
			schema[key] = admitNullableValue(value)
		}
	}
	if nullable, ok := schema["nullable"].(bool); ok && nullable {
		if declared, ok := schema["type"]; ok {
			schema["type"] = withNull(declared)
		}
	}
	return schema
}

func admitNullableValue(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		return admitNullable(typed)
	case []any:
		for i, item := range typed {
			typed[i] = admitNullableValue(item)
		}
		return typed
	}
	return value
}

func withNull(declared any) any {
	switch typed := declared.(type) {
	case string:
		if typed == "null" {
			return typed
		}
		return []any{typed, "null"}
	case []any:
		for _, entry := range typed {
			if name, _ := entry.(string); name == "null" {
				return typed
			}
		}
		return append(typed, "null")
	}
	return declared
}

// ValidateText parses text as exactly one JSON value (trailing text or a second
// value is rejected) and validates it against the compiled schema. Numbers decode
// via json.Number so large integers are not degraded to float64 before validation.
func (c *Compiled) ValidateText(text string) error {
	inst, err := jsonschema.UnmarshalJSON(strings.NewReader(text))
	if err != nil {
		return fmt.Errorf("response is not a single JSON value: %w", err)
	}
	if err := c.schema.Validate(inst); err != nil {
		return fmt.Errorf("response does not match schema: %w", err)
	}
	return nil
}

// Validate compiles schema and validates text, wrapping any failure in a typed
// *llms.ErrStructuredOutputValidation that carries provider/model/choice/stopReason
// and unwraps to the concrete cause. Callers pass the ORIGINAL schema so validation
// reflects exactly what the user asked for, not a provider-transformed variant.
func Validate(schema json.RawMessage, provider, model string, choice int, stopReason, text string) error {
	compiled, err := Compile(schema)
	if err != nil {
		// A schema that fails to compile is a configuration error, not the model
		// answering badly; the preflight normally catches this before the network,
		// but classify it correctly here too rather than blaming the response.
		return fmt.Errorf("%w: schema does not compile: %w", llms.ErrStructuredOutputConfig, err)
	}
	if err := compiled.ValidateText(text); err != nil {
		return wrap(provider, model, choice, stopReason, err)
	}
	return nil
}

func wrap(provider, model string, choice int, stopReason string, cause error) error {
	return &llms.ErrStructuredOutputValidation{
		Provider:   provider,
		Model:      model,
		Choice:     choice,
		StopReason: stopReason,
		Cause:      cause,
	}
}

// ValidateFinalChoices validates every choice that finished normally against the
// schema. normalStop is the vendor's spelling of an ordinary finish: a choice
// that stopped for any other reason, or that answered with a tool call, carries
// no final JSON and is left alone.
func ValidateFinalChoices(
	schema json.RawMessage, provider, model, normalStop string, resp *llms.ContentResponse,
) error {
	if len(schema) == 0 || resp == nil {
		return nil
	}
	for i, choice := range resp.Choices {
		if choice.StopReason != normalStop || len(choice.ToolCalls) > 0 {
			continue
		}
		if err := Validate(schema, provider, model, i, choice.StopReason, choice.Content); err != nil {
			return err
		}
	}
	return nil
}
