package llms

import (
	"encoding/json"
	"errors"
	"fmt"
)

// StructuredOutputConfig requests provider-native, schema-constrained output for a
// single call. Schema is the raw JSON Schema (Draft 2020-12) document, kept
// verbatim: the SDK never rewrites it (no injected required, additionalProperties
// or stripped constraints). Name identifies the schema — mandatory for OpenAI,
// optional for the other providers. Build it through WithStructuredOutput rather
// than by hand so JSONMode and the schema copy stay consistent.
type StructuredOutputConfig struct {
	Name        string          `json:"name,omitempty"`
	Description string          `json:"description,omitempty"`
	Schema      json.RawMessage `json:"schema"`
}

// Clone returns a deep copy whose Schema owns its bytes, so a caller mutating the
// original []byte after building an option cannot alter the stored configuration.
func (c *StructuredOutputConfig) Clone() *StructuredOutputConfig {
	if c == nil {
		return nil
	}
	cp := *c
	if c.Schema != nil {
		cp.Schema = append(json.RawMessage(nil), c.Schema...)
	}
	return &cp
}

// ErrStructuredOutputConfig is the sentinel wrapped by every configuration error
// ValidateStructuredOutput reports (missing JSONMode, empty or non-object schema,
// malformed name). It is detectable without the network.
var ErrStructuredOutputConfig = errors.New("structured output: invalid configuration")

// ErrStructuredOutputUnsupported reports that a model or provider path is KNOWN not
// to accept schema-constrained output. Adapters return it only for documented
// gaps; an unrecognized model is passed through so the provider API is the final
// arbiter, never this local table.
type ErrStructuredOutputUnsupported struct {
	Provider string
	Model    string
	Reason   string
}

func (e *ErrStructuredOutputUnsupported) Error() string {
	msg := "structured output: unsupported"
	if e.Provider != "" {
		msg += " on " + e.Provider
	}
	if e.Model != "" {
		msg += " for model " + e.Model
	}
	if e.Reason != "" {
		msg += ": " + e.Reason
	}
	return msg
}

// ErrStructuredOutputConflict reports that structured output cannot be combined
// with another setting already present on the request (a client-level response
// format, message prefilling, a conflicting response MIME type, and so on).
type ErrStructuredOutputConflict struct {
	Provider string
	Detail   string
}

func (e *ErrStructuredOutputConflict) Error() string {
	msg := "structured output: conflicting configuration"
	if e.Provider != "" {
		msg += " on " + e.Provider
	}
	if e.Detail != "" {
		msg += ": " + e.Detail
	}
	return msg
}

// ErrStructuredOutputValidation is returned when a normal-final response is not a
// single JSON value valid against the requested schema. It carries the provider,
// model, choice index and stop reason for diagnostics and unwraps to the concrete
// cause; it deliberately does not embed the whole model output in its message.
type ErrStructuredOutputValidation struct {
	Provider   string
	Model      string
	Choice     int
	StopReason string
	Cause      error
}

func (e *ErrStructuredOutputValidation) Error() string {
	return fmt.Sprintf("structured output: response validation failed (provider=%s model=%s choice=%d stop_reason=%s): %v",
		e.Provider, e.Model, e.Choice, e.StopReason, e.Cause)
}

func (e *ErrStructuredOutputValidation) Unwrap() error { return e.Cause }

// ValidateStructuredOutput checks the structured-output configuration for
// contradictions detectable without the network: a non-nil config requires
// JSONMode, a non-empty schema whose top-level document is a JSON object, and,
// when a name is set, a value acceptable to the strictest consumer (OpenAI: at
// most 64 chars of ASCII letters, digits, '_' or '-'). It does not verify
// provider-specific schema subsets — that is each adapter's job, and ultimately
// the API's.
func (o *CallOptions) ValidateStructuredOutput() error {
	so := o.StructuredOutput
	if so == nil {
		return nil
	}
	if !o.JSONMode {
		return fmt.Errorf("%w: StructuredOutput set without JSONMode (use WithStructuredOutput)", ErrStructuredOutputConfig)
	}
	if len(so.Schema) == 0 {
		return fmt.Errorf("%w: schema is empty", ErrStructuredOutputConfig)
	}
	// Unmarshalling into a map both proves the document is syntactically valid JSON
	// and that its top level is an object, as every supported provider requires.
	// A literal `null` unmarshals into a nil map without error, so reject it too.
	var doc map[string]json.RawMessage
	if err := json.Unmarshal(so.Schema, &doc); err != nil {
		return fmt.Errorf("%w: schema top level must be a JSON object: %w", ErrStructuredOutputConfig, err)
	}
	if doc == nil {
		return fmt.Errorf("%w: schema top level must be a JSON object, got null", ErrStructuredOutputConfig)
	}
	if so.Name != "" && !isValidStructuredOutputName(so.Name) {
		return fmt.Errorf("%w: name %q must be 1-64 chars of letters, digits, '_' or '-'", ErrStructuredOutputConfig, so.Name)
	}
	return nil
}

// isValidStructuredOutputName reports whether name satisfies the OpenAI schema-name
// rule, the strictest across the supported providers.
func isValidStructuredOutputName(name string) bool {
	if len(name) == 0 || len(name) > 64 {
		return false
	}
	for _, r := range name {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '_', r == '-':
		default:
			return false
		}
	}
	return true
}
