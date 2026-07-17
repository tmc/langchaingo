package openai

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/llms/openai/internal/openaiclient"
)

// newUnitLLM builds an LLM without touching the network; used to exercise the
// request builder and response validator directly.
func newUnitLLM(t *testing.T, opts ...Option) *LLM {
	t.Helper()
	opts = append([]Option{WithToken("unit-test-token")}, opts...)
	llm, err := New(opts...)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return llm
}

func objectSchema() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{"answer":{"type":"string"}},"required":["answer"],"additionalProperties":false}`)
}

func TestCreateChatRequest_ResponseFormatModes(t *testing.T) { //nolint:funlen // table of request-mode sub-cases
	t.Parallel()

	t.Run("WithJSONMode yields json_object", func(t *testing.T) {
		t.Parallel()
		llm := newUnitLLM(t, WithModel("gpt-4o-2024-08-06"))
		var opts llms.CallOptions
		llms.WithJSONMode()(&opts)
		req, err := llm.createChatRequest(nil, opts)
		if err != nil {
			t.Fatal(err)
		}
		if req.ResponseFormat == nil || req.ResponseFormat.Type != "json_object" {
			t.Fatalf("want json_object, got %+v", req.ResponseFormat)
		}
	})

	t.Run("WithStructuredOutput yields json_schema", func(t *testing.T) {
		t.Parallel()
		llm := newUnitLLM(t, WithModel("gpt-4o-2024-08-06"))
		var opts llms.CallOptions
		llms.WithStructuredOutput(llms.StructuredOutputConfig{Name: "s", Description: "d", Schema: objectSchema()})(&opts)
		req, err := llm.createChatRequest(nil, opts)
		if err != nil {
			t.Fatal(err)
		}
		if req.ResponseFormat == nil || req.ResponseFormat.Type != "json_schema" {
			t.Fatalf("want json_schema, got %+v", req.ResponseFormat)
		}
		// The general raw path serializes to json_schema on the wire (it does not
		// populate the public typed JSONSchema field), so assert via marshaling.
		b, err := json.Marshal(req.ResponseFormat)
		if err != nil {
			t.Fatal(err)
		}
		var round struct {
			Type       string `json:"type"`
			JSONSchema struct {
				Name        string `json:"name"`
				Description string `json:"description"`
				Strict      bool   `json:"strict"`
			} `json:"json_schema"`
		}
		if err := json.Unmarshal(b, &round); err != nil {
			t.Fatal(err)
		}
		js := round.JSONSchema
		if js.Name != "s" || js.Description != "d" || !js.Strict {
			t.Fatalf("json_schema fields wrong: %s", b)
		}
	})

	t.Run("conflict with client-level response format", func(t *testing.T) {
		t.Parallel()
		llm := newUnitLLM(t, WithModel("gpt-4o-2024-08-06"),
			WithResponseFormat(&ResponseFormat{Type: "json_object"}))
		var opts llms.CallOptions
		llms.WithStructuredOutput(llms.StructuredOutputConfig{Name: "s", Schema: objectSchema()})(&opts)
		_, err := llm.createChatRequest(nil, opts)
		var conflict *llms.ErrStructuredOutputConflict
		if !errors.As(err, &conflict) {
			t.Fatalf("want ErrStructuredOutputConflict, got %v", err)
		}
	})

	t.Run("missing name is a config error", func(t *testing.T) {
		t.Parallel()
		llm := newUnitLLM(t, WithModel("gpt-4o-2024-08-06"))
		var opts llms.CallOptions
		llms.WithStructuredOutput(llms.StructuredOutputConfig{Schema: objectSchema()})(&opts)
		_, err := llm.createChatRequest(nil, opts)
		if !errors.Is(err, llms.ErrStructuredOutputConfig) {
			t.Fatalf("want ErrStructuredOutputConfig, got %v", err)
		}
	})

	t.Run("known-old model is unsupported", func(t *testing.T) {
		t.Parallel()
		for _, model := range []string{"gpt-3.5-turbo", "gpt-4o-2024-05-13"} {
			llm := newUnitLLM(t, WithModel(model))
			var opts llms.CallOptions
			llms.WithStructuredOutput(llms.StructuredOutputConfig{Name: "s", Schema: objectSchema()})(&opts)
			_, err := llm.createChatRequest(nil, opts)
			var unsup *llms.ErrStructuredOutputUnsupported
			if !errors.As(err, &unsup) {
				t.Fatalf("%s: want ErrStructuredOutputUnsupported, got %v", model, err)
			}
		}
	})

	t.Run("unknown model passes through", func(t *testing.T) {
		t.Parallel()
		llm := newUnitLLM(t, WithModel("gpt-6-ultra-preview"))
		var opts llms.CallOptions
		llms.WithStructuredOutput(llms.StructuredOutputConfig{Name: "s", Schema: objectSchema()})(&opts)
		req, err := llm.createChatRequest(nil, opts)
		if err != nil {
			t.Fatalf("unknown model must pass through, got %v", err)
		}
		if req.ResponseFormat.Type != "json_schema" {
			t.Fatalf("want json_schema, got %+v", req.ResponseFormat)
		}
	})
}

// TestResponseFormatJSONSchema_TypedBuilderStillMarshals guards source and wire
// compatibility of the provider-specific typed builder: the public type keeps its
// original fields, stays readable and comparable, and marshals as before.
func TestResponseFormatJSONSchema_TypedBuilderStillMarshals(t *testing.T) {
	t.Parallel()
	rf := &ResponseFormat{
		Type: "json_schema",
		JSONSchema: &ResponseFormatJSONSchema{
			Name:   "s",
			Strict: true,
			Schema: &ResponseFormatJSONSchemaProperty{
				Type:                 "object",
				Properties:           map[string]*ResponseFormatJSONSchemaProperty{"a": {Type: "string"}},
				AdditionalProperties: false,
				Required:             []string{"a"},
			},
		},
	}
	b, err := json.Marshal(rf)
	if err != nil {
		t.Fatal(err)
	}
	var round map[string]any
	if err := json.Unmarshal(b, &round); err != nil {
		t.Fatal(err)
	}
	js, ok := round["json_schema"].(map[string]any)
	if !ok || js["name"] != "s" || js["strict"] != true {
		t.Fatalf("typed builder marshaled wrong: %s", b)
	}
	if sc, ok := js["schema"].(map[string]any); !ok || sc["type"] != "object" {
		t.Fatalf("schema did not marshal as object: %s", b)
	}
	// Source compatibility: the typed Schema field must remain readable as
	// *ResponseFormatJSONSchemaProperty (not widened to any).
	if got := rf.JSONSchema.Schema.Type; got != "object" {
		t.Errorf("typed Schema must stay readable, Schema.Type = %q", got)
	}
	// Source compatibility: the public type must stay comparable (no []byte field).
	// This block fails to compile if ResponseFormatJSONSchema becomes non-comparable.
	var a, b2 ResponseFormatJSONSchema
	if a != b2 {
		t.Error("zero values must compare equal")
	}
}

// TestRawJSONSchemaResponseFormat covers the general WithStructuredOutput wire path:
// a verbatim schema and optional description are emitted under json_schema without
// going through (or changing) the public typed builder.
func TestRawJSONSchemaResponseFormat(t *testing.T) {
	t.Parallel()
	rf := openaiclient.NewRawJSONSchemaResponseFormat("answer", "an answer",
		json.RawMessage(`{"type":"object","properties":{"x":{"type":"string"}},"required":["x"],"additionalProperties":false}`))

	b, err := json.Marshal(rf)
	if err != nil {
		t.Fatal(err)
	}
	var round map[string]any
	if err := json.Unmarshal(b, &round); err != nil {
		t.Fatal(err)
	}
	if round["type"] != "json_schema" {
		t.Fatalf("type = %v, want json_schema", round["type"])
	}
	js, ok := round["json_schema"].(map[string]any)
	if !ok || js["name"] != "answer" || js["description"] != "an answer" || js["strict"] != true {
		t.Fatalf("raw json_schema fields wrong: %s", b)
	}
	sc, ok := js["schema"].(map[string]any)
	if !ok || sc["additionalProperties"] != false {
		t.Fatalf("raw schema not embedded verbatim: %s", b)
	}
}

func TestValidateOpenAIStructuredSchema_Preflight(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		schema  string
		wantErr bool
	}{
		{"valid root object", `{"type":"object","properties":{"a":{"type":"string"}},"required":["a"],"additionalProperties":false}`, false},
		{"root array rejected", `{"type":"array","items":{"type":"string"}}`, true},
		{"top-level anyOf rejected", `{"anyOf":[{"type":"object"}]}`, true},
		{"missing additionalProperties", `{"type":"object","properties":{"a":{"type":"string"}},"required":["a"]}`, true},
		{"object without properties still needs additionalProperties:false", `{"type":"object"}`, true},
		{"optional property rejected", `{"type":"object","properties":{"a":{"type":"string"},"b":{"type":"string"}},"required":["a"],"additionalProperties":false}`, true},
		{"nested object missing additionalProperties", `{"type":"object","properties":{"a":{"type":"object","properties":{"b":{"type":"string"}},"required":["b"]}},"required":["a"],"additionalProperties":false}`, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := validateOpenAIStructuredSchema(json.RawMessage(tc.schema))
			if tc.wantErr != (err != nil) {
				t.Fatalf("err = %v, wantErr = %v", err, tc.wantErr)
			}
		})
	}
}

func TestValidateStructuredResponse(t *testing.T) {
	t.Parallel()
	llm := newUnitLLM(t, WithModel("gpt-4o-2024-08-06"))
	var opts llms.CallOptions
	llms.WithStructuredOutput(llms.StructuredOutputConfig{Name: "s", Schema: objectSchema()})(&opts)

	mk := func(reason openaiclient.FinishReason, content, refusal string) *openaiclient.ChatCompletionResponse {
		return &openaiclient.ChatCompletionResponse{
			Choices: []*openaiclient.ChatCompletionChoice{
				{FinishReason: reason, Message: openaiclient.ChatMessage{Content: content, Refusal: refusal}},
			},
		}
	}

	t.Run("valid stop passes", func(t *testing.T) {
		t.Parallel()
		if err := llm.validateStructuredResponse(mk(openaiclient.FinishReasonStop, `{"answer":"4"}`, ""), opts); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	t.Run("invalid stop fails with typed error", func(t *testing.T) {
		t.Parallel()
		err := llm.validateStructuredResponse(mk(openaiclient.FinishReasonStop, `{"answer":5}`, ""), opts)
		var ve *llms.ErrStructuredOutputValidation
		if !errors.As(err, &ve) {
			t.Fatalf("want ErrStructuredOutputValidation, got %v", err)
		}
	})
	t.Run("refusal is a typed refusal, not a validation error", func(t *testing.T) {
		t.Parallel()
		err := llm.validateStructuredResponse(mk(openaiclient.FinishReasonStop, "", "I cannot help with that"), opts)
		var refusal *ErrStructuredOutputRefusal
		if !errors.As(err, &refusal) {
			t.Fatalf("want ErrStructuredOutputRefusal, got %v", err)
		}
		var ve *llms.ErrStructuredOutputValidation
		if errors.As(err, &ve) {
			t.Fatal("refusal must not surface as a validation error")
		}
	})
	t.Run("length is not final json", func(t *testing.T) {
		t.Parallel()
		if err := llm.validateStructuredResponse(mk(openaiclient.FinishReasonLength, `{"answer"`, ""), opts); err != nil {
			t.Fatalf("length must be skipped, got %v", err)
		}
	})
	t.Run("content_filter is not final json", func(t *testing.T) {
		t.Parallel()
		if err := llm.validateStructuredResponse(mk(openaiclient.FinishReasonContentFilter, `not json`, ""), opts); err != nil {
			t.Fatalf("content_filter must be skipped, got %v", err)
		}
	})
	t.Run("tool_calls is not final json", func(t *testing.T) {
		t.Parallel()
		if err := llm.validateStructuredResponse(mk(openaiclient.FinishReasonToolCalls, ``, ""), opts); err != nil {
			t.Fatalf("tool_calls must be skipped, got %v", err)
		}
	})
	t.Run("multiple choices report the failing index", func(t *testing.T) {
		t.Parallel()
		result := &openaiclient.ChatCompletionResponse{
			Choices: []*openaiclient.ChatCompletionChoice{
				{FinishReason: openaiclient.FinishReasonStop, Message: openaiclient.ChatMessage{Content: `{"answer":"ok"}`}},
				{FinishReason: openaiclient.FinishReasonStop, Message: openaiclient.ChatMessage{Content: `{"answer":9}`}},
			},
		}
		err := llm.validateStructuredResponse(result, opts)
		var ve *llms.ErrStructuredOutputValidation
		if !errors.As(err, &ve) || ve.Choice != 1 {
			t.Fatalf("want validation error at choice 1, got %v", err)
		}
	})
}
