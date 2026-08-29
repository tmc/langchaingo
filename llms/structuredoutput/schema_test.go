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

func TestRequireClosedObjectsReachesEveryKeywordThatHoldsASchema(t *testing.T) {
	t.Parallel()

	open := `{"type":"object","properties":{"x":{"type":"string"}}}`

	for _, keyword := range []string{
		"contains", "unevaluatedItems", "unevaluatedProperties",
	} {
		t.Run(keyword, func(t *testing.T) {
			t.Parallel()
			schema := []byte(`{"type":"object","additionalProperties":false,"properties":{},"` +
				keyword + `":` + open + `}`)
			if err := structuredoutput.RequireClosedObjects(schema); err == nil {
				t.Fatalf("an object under %q is still an object schema and must be closed", keyword)
			}
		})
	}

	t.Run("dependentSchemas", func(t *testing.T) {
		t.Parallel()
		schema := []byte(`{"type":"object","additionalProperties":false,"properties":{},` +
			`"dependentSchemas":{"x":` + open + `}}`)
		if err := structuredoutput.RequireClosedObjects(schema); err == nil {
			t.Fatal("an object under dependentSchemas must be closed too")
		}
	})
}

func TestAConditionalSchemaIsNotAskedToCloseItsCondition(t *testing.T) {
	t.Parallel()

	for _, keyword := range []string{"not", "if", "then", "else", "propertyNames"} {
		t.Run(keyword, func(t *testing.T) {
			t.Parallel()

			schema := []byte(`{"type":"object","additionalProperties":false,` +
				`"properties":{"kind":{"type":"string"}},"` + keyword +
				`":{"type":"object","properties":{"kind":{"const":"a"}}}}`)
			if err := structuredoutput.RequireClosedObjects(schema); err != nil {
				t.Fatalf("closing a subschema under %q rewrites what the schema accepts: %v", keyword, err)
			}
		})
	}
}

func TestClosingAConditionRewritesWhatTheSchemaAccepts(t *testing.T) {
	t.Parallel()

	base := `{"type":"object","additionalProperties":false,` +
		`"properties":{"kind":{"type":"string"},"x":{"type":"string"},"y":{"type":"string"}},`

	t.Run("a closed if stops matching what the condition was written for", func(t *testing.T) {
		t.Parallel()
		instance := `{"kind":"a","y":"present"}` // matches the condition, missing the required x

		open := mustCompile(t, base+`"if":{"properties":{"kind":{"const":"a"}}},"then":{"required":["x"]}}`)
		if err := open.ValidateText(instance); err == nil {
			t.Fatal("the condition holds, so the missing x must be rejected")
		}

		closed := mustCompile(t, base+
			`"if":{"properties":{"kind":{"const":"a"}},"additionalProperties":false},"then":{"required":["x"]}}`)
		if err := closed.ValidateText(instance); err != nil {
			t.Fatalf("unexpected: %v", err)
		}
	})

	t.Run("a closed then forbids what the root allows", func(t *testing.T) {
		t.Parallel()
		instance := `{"kind":"a","x":"present","y":"allowed by the root"}`

		open := mustCompile(t, base+`"if":{"properties":{"kind":{"const":"a"}}},`+
			`"then":{"type":"object","properties":{"x":{"type":"string"}},"required":["x"]}}`)
		if err := open.ValidateText(instance); err != nil {
			t.Fatalf("the instance satisfies the schema: %v", err)
		}

		closed := mustCompile(t, base+`"if":{"properties":{"kind":{"const":"a"}}},`+
			`"then":{"type":"object","additionalProperties":false,`+
			`"properties":{"x":{"type":"string"}},"required":["x"]}}`)
		if err := closed.ValidateText(instance); err == nil {
			t.Fatal("closing then rejected an instance the root allows")
		}
	})

	t.Run("a closed not narrows what is excluded", func(t *testing.T) {
		t.Parallel()
		instance := `{"kind":"forbidden","y":"v"}`

		open := mustCompile(t, base+`"not":{"properties":{"kind":{"const":"forbidden"}}}}`)
		if err := open.ValidateText(instance); err == nil {
			t.Fatal("the excluded shape must be rejected")
		}

		closed := mustCompile(t, base+`"not":{"additionalProperties":false,`+
			`"properties":{"kind":{"const":"forbidden"}}}}`)
		if err := closed.ValidateText(instance); err != nil {
			t.Fatalf("unexpected: %v", err)
		}
	})
}

func mustCompile(t *testing.T, schema string) *structuredoutput.Compiled {
	t.Helper()
	c, err := structuredoutput.Compile(json.RawMessage(schema))
	if err != nil {
		t.Fatalf("compile: %v", err)
	}
	return c
}
