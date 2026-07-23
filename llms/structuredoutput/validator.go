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
	const resource = "structuredoutput:schema"
	comp := jsonschema.NewCompiler()
	if err := comp.AddResource(resource, doc); err != nil {
		return nil, fmt.Errorf("load schema: %w", err)
	}
	sch, err := comp.Compile(resource)
	if err != nil {
		return nil, fmt.Errorf("compile schema: %w", err)
	}
	return &Compiled{schema: sch}, nil
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
