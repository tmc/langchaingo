// Package structuredoutput compiles a JSON Schema and validates a model's final
// text response against it, backing the provider-neutral llms.WithStructuredOutput
// contract.
//
// It is kept separate from the base llms package so that package does not depend on
// an external JSON Schema library: this package imports llms (for the typed
// llms.ErrStructuredOutputValidation), llms does not import this package, and
// provider adapters may import both without an import cycle.
//
// Validate compiles a schema (JSON Schema Draft 2020-12, via
// github.com/santhosh-tekuri/jsonschema/v6) and checks that the response text is
// exactly one JSON value (trailing text or a second value is rejected) matching it.
// Numbers decode with json.Number so large integers survive; enum/const stay
// case-sensitive. Failures are wrapped in *llms.ErrStructuredOutputValidation with
// the provider, model, choice index and stop reason.
//
// Adapters call it only for a normal-final response; refusals, tool-use turns and
// truncated/blocked responses are never validated as final JSON.
package structuredoutput
