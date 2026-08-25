package structuredoutput_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/llms/structuredoutput"
)

func TestRequireClosedObjects(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		schema  string
		wantErr bool
	}{
		{"closed root object", `{"type":"object","properties":{"a":{"type":"string"}},"required":["a"],"additionalProperties":false}`, false},
		{"root object missing additionalProperties", `{"type":"object","properties":{"a":{"type":"string"}},"required":["a"]}`, true},
		{"root object additionalProperties true", `{"type":"object","properties":{"a":{"type":"string"}},"additionalProperties":true}`, true},
		{"empty object without additionalProperties", `{"type":"object"}`, true},
		{"nested object missing additionalProperties", `{"type":"object","properties":{"a":{"type":"object","properties":{"b":{"type":"string"}}}},"additionalProperties":false}`, true},
		{"nested object closed", `{"type":"object","properties":{"a":{"type":"object","properties":{"b":{"type":"string"}},"additionalProperties":false}},"additionalProperties":false}`, false},
		{"object inside array items", `{"type":"array","items":{"type":"object","properties":{"a":{"type":"string"}}}}`, true},
		{"non-object schema is fine", `{"type":"string"}`, false},
		{"array of strings is fine", `{"type":"array","items":{"type":"string"}}`, false},
		{"union type naming object", `{"type":["object","null"]}`, true},
		{"union type naming object, closed", `{"type":["object","null"],"additionalProperties":false}`, false},
		{"union type without object is fine", `{"type":["string","null"]}`, false},
		{"object under prefixItems", `{"type":"array","prefixItems":[{"type":"object","properties":{"a":{"type":"string"}}}],"items":false}`, true},
		{"object under prefixItems, closed", `{"type":"array","prefixItems":[{"type":"object","properties":{"a":{"type":"string"}},"additionalProperties":false}],"items":false}`, false},
		{"object under tuple-form items", `{"type":"array","items":[{"type":"object","properties":{"a":{"type":"string"}}}]}`, true},
		{"object under patternProperties", `{"type":"object","additionalProperties":false,"patternProperties":{"^k":{"type":"object","properties":{"a":{"type":"string"}}}}}`, true},
		{"object under patternProperties, closed", `{"type":"object","additionalProperties":false,"patternProperties":{"^k":{"type":"object","properties":{"a":{"type":"string"}},"additionalProperties":false}}}`, false},
		{"object under a schema-valued additionalProperties", `{"additionalProperties":{"type":"object","properties":{"a":{"type":"string"}}}}`, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := structuredoutput.RequireClosedObjects(json.RawMessage(tc.schema))
			if tc.wantErr != (err != nil) {
				t.Fatalf("RequireClosedObjects error = %v, wantErr = %v", err, tc.wantErr)
			}
			if tc.wantErr && !errors.Is(err, llms.ErrStructuredOutputConfig) {
				t.Errorf("error must wrap ErrStructuredOutputConfig, got %v", err)
			}
		})
	}
}
