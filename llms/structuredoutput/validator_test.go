package structuredoutput_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/llms/structuredoutput"
)

func TestValidate(t *testing.T) {
	t.Parallel()

	objectSchema := json.RawMessage(`{
		"type": "object",
		"properties": {"name": {"type": "string"}, "age": {"type": "integer"}},
		"required": ["name", "age"],
		"additionalProperties": false
	}`)

	cases := []struct {
		name    string
		schema  json.RawMessage
		text    string
		wantErr bool
	}{
		{"valid object", objectSchema, `{"name":"a","age":3}`, false},
		{"valid array", json.RawMessage(`{"type":"array","items":{"type":"number"}}`), `[1,2,3]`, false},
		{"valid primitive", json.RawMessage(`{"type":"string"}`), `"hello"`, false},
		{"syntactically invalid json", objectSchema, `{"name":`, true},
		{"trailing text", objectSchema, `{"name":"a","age":3} trailing`, true},
		{"second json value", objectSchema, `{"name":"a","age":3}{"x":1}`, true},
		{"missing required", objectSchema, `{"name":"a"}`, true},
		{"additional property", objectSchema, `{"name":"a","age":3,"extra":1}`, true},
		{"wrong type", objectSchema, `{"name":"a","age":"three"}`, true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := structuredoutput.Validate(tc.schema, "openai", "gpt-x", 0, "stop", tc.text)
			if tc.wantErr != (err != nil) {
				t.Fatalf("Validate error = %v, wantErr = %v", err, tc.wantErr)
			}
		})
	}
}

func TestValidate_EnumIsCaseSensitive(t *testing.T) {
	t.Parallel()
	schema := json.RawMessage(`{"type":"object","properties":{"color":{"enum":["Red","Green"]}},"required":["color"],"additionalProperties":false}`)
	if err := structuredoutput.Validate(schema, "anthropic", "claude", 0, "end_turn", `{"color":"red"}`); err == nil {
		t.Error("case mismatch on enum must fail: schema enum is case-sensitive")
	}
	if err := structuredoutput.Validate(schema, "anthropic", "claude", 0, "end_turn", `{"color":"Red"}`); err != nil {
		t.Errorf("exact enum case must pass: %v", err)
	}
}

func TestValidate_Draft2020Ref(t *testing.T) {
	t.Parallel()
	// $defs + local $ref is a Draft 2020-12 construct the validator must support.
	schema := json.RawMessage(`{
		"type": "object",
		"properties": {"a": {"$ref": "#/$defs/pos"}},
		"required": ["a"],
		"additionalProperties": false,
		"$defs": {"pos": {"type": "integer", "minimum": 1}}
	}`)
	if err := structuredoutput.Validate(schema, "openai", "gpt-x", 0, "stop", `{"a":5}`); err != nil {
		t.Errorf("valid $ref instance must pass: %v", err)
	}
	if err := structuredoutput.Validate(schema, "openai", "gpt-x", 0, "stop", `{"a":0}`); err == nil {
		t.Error("instance violating referenced schema must fail")
	}
}

func TestValidate_TypedError(t *testing.T) {
	t.Parallel()
	schema := json.RawMessage(`{"type":"object","properties":{"x":{"type":"integer"}},"required":["x"],"additionalProperties":false}`)
	err := structuredoutput.Validate(schema, "bedrock", "claude-x", 2, "end_turn", `{"x":"no"}`)
	if err == nil {
		t.Fatal("expected validation error")
	}
	var ve *llms.ErrStructuredOutputValidation
	if !errors.As(err, &ve) {
		t.Fatalf("error must be *llms.ErrStructuredOutputValidation, got %T", err)
	}
	if ve.Provider != "bedrock" || ve.Model != "claude-x" || ve.Choice != 2 || ve.StopReason != "end_turn" {
		t.Errorf("typed error lost context: %+v", ve)
	}
	if ve.Unwrap() == nil {
		t.Error("typed error must retain an unwrap-able cause")
	}
}

func TestCompileDoesNotReadTheFilesystem(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	present := filepath.Join(dir, "present.json")
	if err := os.WriteFile(present, []byte(`{"type":"integer","minimum":42}`), 0o600); err != nil {
		t.Fatal(err)
	}
	absent := filepath.Join(dir, "absent.json")

	refSchema := func(target string) json.RawMessage {
		return json.RawMessage(fmt.Sprintf(
			`{"type":"object","additionalProperties":false,"properties":{"n":{"$ref":%q}}}`,
			"file://"+target))
	}

	_, presentErr := structuredoutput.Compile(refSchema(present))
	_, absentErr := structuredoutput.Compile(refSchema(absent))

	if presentErr == nil {
		t.Fatal("a file:// reference must not compile: the schema would be read off the caller's disk")
	}
	if absentErr == nil {
		t.Fatal("a file:// reference must not compile")
	}
	const refused = "no URLLoader registered"
	for label, err := range map[string]error{"present": presentErr, "absent": absentErr} {
		if !strings.Contains(err.Error(), refused) {
			t.Errorf("%s path: want the reference refused before any filesystem access, got %v", label, err)
		}
	}
}

func TestCompileStillAcceptsInternalAndDraftRefs(t *testing.T) {
	t.Parallel()

	for name, schema := range map[string]string{
		"internal $defs": `{"$defs":{"leaf":{"type":"string"}},"type":"object",` +
			`"additionalProperties":false,"properties":{"a":{"$ref":"#/$defs/leaf"}}}`,
		"explicit draft 2020-12": `{"$schema":"https://json-schema.org/draft/2020-12/schema",` +
			`"type":"object","additionalProperties":false,"properties":{"a":{"type":"string"}}}`,
		"explicit draft-07": `{"$schema":"http://json-schema.org/draft-07/schema#","type":"string"}`,
	} {
		if _, err := structuredoutput.Compile(json.RawMessage(schema)); err != nil {
			t.Errorf("%s: %v", name, err)
		}
	}
}

func TestTheValidatorAdmitsTheNullableSpellingTheDoorsSendOnTheWire(t *testing.T) {
	t.Parallel()

	schema := json.RawMessage(`{"type":"object","additionalProperties":false,` +
		`"properties":{"a":{"type":"string","nullable":true}},"required":["a"]}`)

	compiled, err := structuredoutput.Compile(schema)
	if err != nil {
		t.Fatalf("Compile() error: %v", err)
	}

	if err := compiled.ValidateText(`{"a":null}`); err != nil {
		t.Fatalf("a null the schema declared nullable must validate, got: %v", err)
	}
	if err := compiled.ValidateText(`{"a":"x"}`); err != nil {
		t.Fatalf("the declared type must still validate, got: %v", err)
	}
	if err := compiled.ValidateText(`{"a":7}`); err == nil {
		t.Fatal("nullable widens the type by null alone, not by every type")
	}
}

func TestNullableWithoutADeclaredTypeStillCompiles(t *testing.T) {
	t.Parallel()

	schema := json.RawMessage(`{"type":"object","additionalProperties":false,` +
		`"properties":{"a":{"nullable":true}},"required":["a"]}`)

	compiled, err := structuredoutput.Compile(schema)
	if err != nil {
		t.Fatalf("Compile() error: %v", err)
	}
	for _, text := range []string{`{"a":null}`, `{"a":"x"}`, `{"a":7}`} {
		if err := compiled.ValidateText(text); err != nil {
			t.Fatalf("an undeclared type constrains nothing, %s must validate: %v", text, err)
		}
	}
}

func TestNullableLeavesInstanceLiteralsAlone(t *testing.T) {
	t.Parallel()

	schema := json.RawMessage(`{"type":"object","additionalProperties":false,` +
		`"properties":{"a":{"enum":[{"nullable":true,"type":"widget"},"plain"]},` +
		`"b":{"const":{"nullable":true,"type":"widget"}}},` +
		`"required":["a","b"]}`)

	compiled, err := structuredoutput.Compile(schema)
	if err != nil {
		t.Fatalf("Compile() error: %v", err)
	}
	literal := `{"nullable":true,"type":"widget"}`
	if err := compiled.ValidateText(`{"a":` + literal + `,"b":` + literal + `}`); err != nil {
		t.Fatalf("an enum or const literal is data, not a schema to widen: %v", err)
	}
}

func TestNullableReachesNestedSubschemas(t *testing.T) {
	t.Parallel()

	schema := json.RawMessage(`{"type":"object","additionalProperties":false,` +
		`"properties":{"list":{"type":"array","items":{"type":"string","nullable":true}},` +
		`"pick":{"anyOf":[{"type":"integer","nullable":true}]},` +
		`"ref":{"$ref":"#/$defs/leaf"}},` +
		`"$defs":{"leaf":{"type":"string","nullable":true}},` +
		`"required":["list","pick","ref"]}`)

	compiled, err := structuredoutput.Compile(schema)
	if err != nil {
		t.Fatalf("Compile() error: %v", err)
	}
	if err := compiled.ValidateText(`{"list":[null],"pick":null,"ref":null}`); err != nil {
		t.Fatalf("nullable must be admitted inside items, anyOf and $defs: %v", err)
	}
	if err := compiled.ValidateText(`{"list":[1],"pick":null,"ref":null}`); err == nil {
		t.Fatal("nullable widens by null alone, the item type still holds")
	}
}
