package vertex

import (
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"

	"cloud.google.com/go/vertexai/genai"

	"github.com/vxcontrol/langchaingo/llms"
)

func TestVertex_12F0_HandMaintained(t *testing.T) {
	t.Parallel()
	src, err := os.ReadFile("vertex.go")
	if err != nil {
		t.Fatal(err)
	}
	s := string(src)
	if strings.Contains(s, "automatically generated") {
		t.Error("vertex.go must no longer claim to be automatically generated")
	}
	if !strings.Contains(s, "hand-maintained") {
		t.Error("vertex.go must document that it is hand-maintained")
	}
}

func TestConvertJSONSchemaToVertex(t *testing.T) { //nolint:funlen // table-driven test
	t.Parallel()

	t.Run("representable object", func(t *testing.T) {
		t.Parallel()
		raw := json.RawMessage(`{
			"type":"object",
			"properties":{
				"name":{"type":"string"},
				"tags":{"type":"array","items":{"type":"string"}},
				"age":{"type":"integer"},
				"color":{"type":"string","enum":["red","green"]}
			},
			"required":["name"],
			"additionalProperties":false
		}`)
		schema, err := convertJSONSchemaToVertex(raw)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if schema.Type != genai.TypeObject {
			t.Errorf("root type = %v, want object", schema.Type)
		}
		if len(schema.Properties) != 4 {
			t.Errorf("properties = %d, want 4", len(schema.Properties))
		}
		if schema.Properties["tags"].Type != genai.TypeArray || schema.Properties["tags"].Items.Type != genai.TypeString {
			t.Error("array items not converted")
		}
		if len(schema.Properties["color"].Enum) != 2 {
			t.Error("enum not converted")
		}
		if len(schema.Required) != 1 || schema.Required[0] != "name" {
			t.Errorf("required = %v", schema.Required)
		}
	})

	t.Run("string constraints are preserved", func(t *testing.T) {
		t.Parallel()
		schema, err := convertJSONSchemaToVertex(json.RawMessage(`{"type":"string","minLength":2,"maxLength":8,"pattern":"^[a-z]+$"}`))
		if err != nil {
			t.Fatal(err)
		}
		if schema.MinLength != 2 || schema.MaxLength != 8 || schema.Pattern != "^[a-z]+$" {
			t.Errorf("string constraints lost: minLen=%d maxLen=%d pattern=%q", schema.MinLength, schema.MaxLength, schema.Pattern)
		}
	})

	t.Run("array/numeric constraints are preserved", func(t *testing.T) {
		t.Parallel()
		schema, err := convertJSONSchemaToVertex(json.RawMessage(`{"type":"array","items":{"type":"integer","minimum":1,"maximum":9},"minItems":1,"maxItems":3}`))
		if err != nil {
			t.Fatal(err)
		}
		if schema.MinItems != 1 || schema.MaxItems != 3 {
			t.Errorf("array constraints lost: min=%d max=%d", schema.MinItems, schema.MaxItems)
		}
		if schema.Items.Minimum != 1 || schema.Items.Maximum != 9 {
			t.Errorf("numeric constraints lost: min=%v max=%v", schema.Items.Minimum, schema.Items.Maximum)
		}
	})

	t.Run("unknown keyword is rejected", func(t *testing.T) {
		t.Parallel()
		_, err := convertJSONSchemaToVertex(json.RawMessage(`{"type":"integer","multipleOf":2}`))
		var unsup *llms.ErrStructuredOutputUnsupported
		if !errors.As(err, &unsup) {
			t.Fatalf("unknown constraint must be rejected, got %v", err)
		}
	})

	t.Run("fractional integer keyword is rejected", func(t *testing.T) {
		t.Parallel()
		_, err := convertJSONSchemaToVertex(json.RawMessage(`{"type":"array","items":{"type":"string"},"minItems":1.5}`))
		var unsup *llms.ErrStructuredOutputUnsupported
		if !errors.As(err, &unsup) {
			t.Fatalf("fractional minItems must be rejected, got %v", err)
		}
	})

	t.Run("required with non-string is rejected", func(t *testing.T) {
		t.Parallel()
		_, err := convertJSONSchemaToVertex(json.RawMessage(`{"type":"object","properties":{"a":{"type":"string"}},"required":[1],"additionalProperties":false}`))
		var unsup *llms.ErrStructuredOutputUnsupported
		if !errors.As(err, &unsup) {
			t.Fatalf("non-string required must be rejected, got %v", err)
		}
	})

	t.Run("nullable via [T,null]", func(t *testing.T) {
		t.Parallel()
		schema, err := convertJSONSchemaToVertex(json.RawMessage(`{"type":["string","null"]}`))
		if err != nil {
			t.Fatal(err)
		}
		if schema.Type != genai.TypeString || !schema.Nullable {
			t.Errorf("want nullable string, got type=%v nullable=%v", schema.Type, schema.Nullable)
		}
	})

	unrepresentable := map[string]string{
		"$ref":                      `{"$ref":"#/x"}`,
		"anyOf":                     `{"anyOf":[{"type":"string"}]}`,
		"oneOf":                     `{"oneOf":[{"type":"string"}]}`,
		"const":                     `{"const":"x"}`,
		"additionalProperties true": `{"type":"object","properties":{"a":{"type":"string"}},"additionalProperties":true}`,
		"union type":                `{"type":["string","number"]}`,
	}
	for name, raw := range unrepresentable {
		t.Run("rejects "+name, func(t *testing.T) {
			t.Parallel()
			_, err := convertJSONSchemaToVertex(json.RawMessage(raw))
			var unsup *llms.ErrStructuredOutputUnsupported
			if !errors.As(err, &unsup) {
				t.Fatalf("want ErrStructuredOutputUnsupported, got %v", err)
			}
		})
	}
}

func TestApplyVertexResponseFormat(t *testing.T) {
	t.Parallel()
	schema := &llms.StructuredOutputConfig{Name: "s", Schema: json.RawMessage(`{"type":"object","properties":{"a":{"type":"string"}},"required":["a"],"additionalProperties":false}`)}

	t.Run("structured output sets ResponseSchema and MIME", func(t *testing.T) {
		t.Parallel()
		model := &genai.GenerativeModel{}
		opts := &llms.CallOptions{JSONMode: true, StructuredOutput: schema}
		if err := applyVertexResponseFormat(model, opts); err != nil {
			t.Fatal(err)
		}
		if model.ResponseMIMEType != ResponseMIMETypeJson {
			t.Errorf("MIME = %q", model.ResponseMIMEType)
		}
		if model.ResponseSchema == nil {
			t.Error("ResponseSchema must be set")
		}
	})

	t.Run("other MIME conflicts", func(t *testing.T) {
		t.Parallel()
		model := &genai.GenerativeModel{}
		mime := "text/plain"
		opts := &llms.CallOptions{JSONMode: true, ResponseMIMEType: &mime, StructuredOutput: schema}
		err := applyVertexResponseFormat(model, opts)
		var conflict *llms.ErrStructuredOutputConflict
		if !errors.As(err, &conflict) {
			t.Fatalf("want conflict, got %v", err)
		}
	})

	t.Run("schema-less JSONMode unchanged", func(t *testing.T) {
		t.Parallel()
		model := &genai.GenerativeModel{}
		if err := applyVertexResponseFormat(model, &llms.CallOptions{JSONMode: true}); err != nil {
			t.Fatal(err)
		}
		if model.ResponseMIMEType != ResponseMIMETypeJson || model.ResponseSchema != nil {
			t.Error("schema-less mode must set only MIME, not ResponseSchema")
		}
	})
}

func TestValidateVertexStructuredOutput(t *testing.T) {
	t.Parallel()
	opts := &llms.CallOptions{
		JSONMode:         true,
		StructuredOutput: &llms.StructuredOutputConfig{Name: "s", Schema: json.RawMessage(`{"type":"object","properties":{"a":{"type":"string"}},"required":["a"],"additionalProperties":false}`)},
	}
	stop := genai.FinishReasonStop.String()

	if err := validateVertexStructuredOutput(opts, &llms.ContentResponse{Choices: []*llms.ContentChoice{{Content: `{"a":"ok"}`, StopReason: stop}}}); err != nil {
		t.Errorf("valid STOP must pass: %v", err)
	}

	err := validateVertexStructuredOutput(opts, &llms.ContentResponse{Choices: []*llms.ContentChoice{{Content: `{"a":1}`, StopReason: stop}}})
	var ve *llms.ErrStructuredOutputValidation
	if !errors.As(err, &ve) {
		t.Errorf("invalid STOP must fail typed: %v", err)
	}
}
