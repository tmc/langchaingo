package openai

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/llms/openai/internal/openaiclient"
	"github.com/vxcontrol/langchaingo/llms/structuredoutput"
)

const providerOpenAI = "openai"

// ErrStructuredOutputRefusal reports that the model declined a structured-output
// request (OpenAI Structured Outputs). A refusal may legitimately not match the
// schema, so it is a distinct typed outcome rather than a validation failure. The
// usage-carrying ContentResponse is returned alongside this error.
type ErrStructuredOutputRefusal struct {
	Model   string
	Choice  int
	Refusal string
}

func (e *ErrStructuredOutputRefusal) Error() string {
	return fmt.Sprintf("openai structured output: model refused (model=%s choice=%d): %s", e.Model, e.Choice, e.Refusal)
}

// setStructuredOutput translates a per-call llms.StructuredOutput into OpenAI's
// json_schema response format with strict:true. It takes precedence over the
// schema-less JSONMode json_object and returns a typed conflict against a
// client-level response format rather than silently overwriting one.
func (o *LLM) setStructuredOutput(req *openaiclient.ChatRequest, opts llms.CallOptions) error {
	so := opts.StructuredOutput
	if so == nil {
		return nil
	}
	if err := opts.ValidateStructuredOutput(); err != nil {
		return err
	}
	if so.Name == "" {
		return fmt.Errorf("%w: openai structured output requires a schema name", llms.ErrStructuredOutputConfig)
	}
	if o.client.ResponseFormat != nil {
		return &llms.ErrStructuredOutputConflict{
			Provider: providerOpenAI,
			Detail:   "per-call WithStructuredOutput conflicts with client-level WithResponseFormat",
		}
	}
	model := o.effectiveModel(opts)
	if openAIStructuredOutputUnsupported(model) {
		return &llms.ErrStructuredOutputUnsupported{
			Provider: providerOpenAI,
			Model:    model,
			Reason:   "model predates Structured Outputs (json_schema)",
		}
	}
	if err := validateOpenAIStructuredSchema(so.Schema); err != nil {
		return err
	}
	req.ResponseFormat = &openaiclient.ResponseFormat{
		Type: "json_schema",
		JSONSchema: &openaiclient.ResponseFormatJSONSchema{
			Name:        so.Name,
			Description: so.Description,
			Strict:      true,
			SchemaRaw:   so.Schema,
		},
	}
	return nil
}

// validateStructuredResponse checks each normal-final ("stop") choice against the
// requested schema. A refusal is surfaced through GenerationInfo, not validated;
// length, content_filter, tool_calls and function_call are not final JSON either.
func (o *LLM) validateStructuredResponse(result *openaiclient.ChatCompletionResponse, opts llms.CallOptions) error {
	so := opts.StructuredOutput
	if so == nil {
		return nil
	}
	model := o.effectiveModel(opts)
	for i, c := range result.Choices {
		// A refusal is a distinct typed outcome, never a JSON-validation error.
		if c.Message.Refusal != "" {
			return &ErrStructuredOutputRefusal{Model: model, Choice: i, Refusal: c.Message.Refusal}
		}
		if c.FinishReason != openaiclient.FinishReasonStop {
			continue
		}
		if err := structuredoutput.Validate(so.Schema, providerOpenAI, model, i, string(c.FinishReason), c.Message.Content); err != nil {
			return err
		}
	}
	return nil
}

// openAIStructuredOutputUnsupported reports models KNOWN to lack Structured Outputs
// (json_schema). Unknown or newer names pass through so the local table never
// blocks a future model — the API is the final arbiter.
func openAIStructuredOutputUnsupported(model string) bool {
	m := strings.ToLower(model)
	if idx := strings.LastIndex(m, "/"); idx != -1 {
		m = m[idx+1:]
	}
	switch {
	case strings.HasPrefix(m, "gpt-3.5"):
		return true
	case m == "gpt-4", strings.HasPrefix(m, "gpt-4-0"), strings.HasPrefix(m, "gpt-4-32k"), strings.HasPrefix(m, "gpt-4-turbo"):
		return true
	case m == "gpt-4o-2024-05-13":
		// The first gpt-4o snapshot predates json_schema (added in 2024-08-06).
		return true
	default:
		return false
	}
}

// validateOpenAIStructuredSchema enforces the OpenAI Structured Outputs subset that
// is checkable locally: a root object, no top-level anyOf, and every object node
// setting additionalProperties:false with all its properties listed in required.
func validateOpenAIStructuredSchema(raw json.RawMessage) error {
	var root map[string]any
	if err := json.Unmarshal(raw, &root); err != nil {
		return fmt.Errorf("%w: schema must be a JSON object: %w", llms.ErrStructuredOutputConfig, err)
	}
	if _, ok := root["anyOf"]; ok {
		return fmt.Errorf("%w: OpenAI does not allow anyOf at the schema root", llms.ErrStructuredOutputConfig)
	}
	if t, _ := root["type"].(string); t != "object" {
		return fmt.Errorf("%w: OpenAI structured output requires a root object schema", llms.ErrStructuredOutputConfig)
	}
	return checkOpenAIObjectNodes(root)
}

func checkOpenAIObjectNodes(node map[string]any) error {
	props, hasProps := node["properties"].(map[string]any)
	// An object node is any schema whose type is "object" OR that carries
	// properties. OpenAI requires additionalProperties:false on every object,
	// including one with no declared properties (e.g. {"type":"object"}).
	if t, _ := node["type"].(string); t == "object" || hasProps {
		if ap, ok := node["additionalProperties"].(bool); !ok || ap {
			return fmt.Errorf("%w: every object schema must set additionalProperties:false", llms.ErrStructuredOutputConfig)
		}
	}
	if hasProps {
		required := map[string]bool{}
		if reqs, ok := node["required"].([]any); ok {
			for _, r := range reqs {
				if s, ok := r.(string); ok {
					required[s] = true
				}
			}
		}
		for name := range props {
			if !required[name] {
				return fmt.Errorf("%w: property %q must be listed in required (OpenAI strict)", llms.ErrStructuredOutputConfig, name)
			}
		}
	}
	for _, child := range openAISubschemas(node) {
		if m, ok := child.(map[string]any); ok {
			if err := checkOpenAIObjectNodes(m); err != nil {
				return err
			}
		}
	}
	return nil
}

// openAISubschemas returns nested schema nodes worth recursing into for the
// object-node checks (properties, $defs/definitions, items, composition arrays).
func openAISubschemas(node map[string]any) []any {
	var out []any
	if props, ok := node["properties"].(map[string]any); ok {
		for _, v := range props {
			out = append(out, v)
		}
	}
	for _, key := range []string{"$defs", "definitions"} {
		if defs, ok := node[key].(map[string]any); ok {
			for _, v := range defs {
				out = append(out, v)
			}
		}
	}
	if items, ok := node["items"].(map[string]any); ok {
		out = append(out, items)
	}
	for _, key := range []string{"anyOf", "oneOf", "allOf"} {
		if arr, ok := node[key].([]any); ok {
			out = append(out, arr...)
		}
	}
	return out
}
