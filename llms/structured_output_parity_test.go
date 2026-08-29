package llms_test

import (
	"encoding/json"
	"testing"

	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/llms/structuredoutput"
)

func TestThePreflightAcceptsExactlyWhatTheResponsePathCompiles(t *testing.T) {
	t.Parallel()

	for _, schema := range []string{
		`{"type":"object","additionalProperties":false,"properties":{"a":{"type":"string","nullable":true}}}`,
		`{"type":"object","additionalProperties":false,"properties":{"a":{"nullable":true}}}`,
		`{"type":"object","additionalProperties":false,"properties":{"a":{"type":["string"],"nullable":true}}}`,
		`{"type":"object","additionalProperties":false,"properties":{"a":{"type":"null","nullable":true}}}`,
		`{"type":"object","additionalProperties":false,"properties":{"a":{"type":{"bad":1},"nullable":true}}}`,
		`{"type":"object","additionalProperties":false,"properties":{"a":{"enum":[{"nullable":true,"type":"w"}]}}}`,
		`{"type":"object","additionalProperties":false,"properties":{"a":{"type":"string","pattern":"(?=x)"}}}`,
		`{"type":"object","additionalProperties":false,"$defs":{"leaf":{"type":"string","nullable":true}},` +
			`"properties":{"a":{"$ref":"#/$defs/leaf"}}}`,
		`{"type":"object","additionalProperties":false,"properties":{"a":{"type":"integer","nullable":true,"minimum":"nope"}}}`,
	} {
		t.Run(schema, func(t *testing.T) {
			t.Parallel()

			opts := &llms.CallOptions{
				JSONMode:         true,
				StructuredOutput: &llms.StructuredOutputConfig{Schema: json.RawMessage(schema)},
			}
			preflight := opts.ValidateStructuredOutput()
			_, response := structuredoutput.Compile(json.RawMessage(schema))

			if (preflight == nil) != (response == nil) {
				t.Fatalf("preflight and response path disagree: preflight=%v response=%v", preflight, response)
			}
		})
	}
}
