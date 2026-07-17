package bedrockclient

import (
	"reflect"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/llms/reasoning"
	"github.com/vxcontrol/langchaingo/llms/structuredoutput"
)

const providerBedrock = "bedrock"

// isSmithyValidObject validates that an object contains only simple types
// or structs with proper document tags on all public fields
func isSmithyValidObject(value any) bool {
	if value == nil {
		return true
	}

	v := reflect.ValueOf(value)
	t := reflect.TypeOf(value)

	if isSimpleType(t) {
		return true
	}

	if t.Kind() == reflect.Ptr {
		if v.IsNil() {
			return true
		}
		return isSmithyValidObject(v.Elem().Interface())
	}

	if t.Kind() == reflect.Struct {
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)

			if !field.IsExported() {
				continue
			}

			if _, hasTag := field.Tag.Lookup("document"); !hasTag {
				return false
			}

			fieldValue := v.Field(i).Interface()
			if !isSmithyValidObject(fieldValue) {
				return false
			}
		}
		return true
	}

	if t.Kind() == reflect.Slice {
		for i := 0; i < v.Len(); i++ {
			elementValue := v.Index(i).Interface()
			if !isSmithyValidObject(elementValue) {
				return false
			}
		}
		return true
	}

	if t.Kind() == reflect.Map {
		for _, key := range v.MapKeys() {
			// Check key validity
			if !isSmithyValidObject(key.Interface()) {
				return false
			}
			// Check value validity
			mapValue := v.MapIndex(key).Interface()
			if !isSmithyValidObject(mapValue) {
				return false
			}
		}
		return true
	}

	return false
}

func isSimpleType(t reflect.Type) bool {
	switch t.Kind() {
	case reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64,
		reflect.Complex64, reflect.Complex128,
		reflect.String:
		return true
	default:
		return false
	}
}

// applyConverseStructuredOutput sets the native Converse OutputConfig.TextFormat
// from a per-call schema. Converse Structured Outputs is not Claude-only, so there
// is no local model gate here — an unsupported model surfaces as the provider 4xx.
// The AWS JsonSchemaDefinition.Schema field is a JSON string, so the raw schema is
// passed as-is with no second encoding.
func applyConverseStructuredOutput(input *ConverseInput, converseInput *bedrockruntime.ConverseInput) error {
	so := input.StructuredOutput
	if so == nil {
		return nil
	}
	// Bedrock rejects an object schema that omits additionalProperties:false;
	// surface that documented, locally-detectable requirement as a typed error.
	if err := structuredoutput.RequireClosedObjects(so.Schema); err != nil {
		return err
	}
	converseInput.OutputConfig = &types.OutputConfig{
		TextFormat: &types.OutputFormat{
			Type: types.OutputFormatTypeJsonSchema,
			Structure: &types.OutputFormatStructureMemberJsonSchema{
				Value: types.JsonSchemaDefinition{
					Schema:      aws.String(string(so.Schema)),
					Name:        ptrStringOrNil(so.Name),
					Description: ptrStringOrNil(so.Description),
				},
			},
		},
	}
	return nil
}

// applyAnthropicStructuredOutput folds a per-call schema into the legacy Anthropic
// output_config.format, preserving any effort already set. A known-legacy model that
// predates structured output is rejected with a typed error before the request.
func applyAnthropicStructuredOutput(input *anthropicTextGenerationInput, modelID string, so *llms.StructuredOutputConfig) error {
	if so == nil {
		return nil
	}
	if !reasoning.ClaudeSupportsStructuredOutput(modelID) {
		return &llms.ErrStructuredOutputUnsupported{
			Provider: providerBedrock,
			Model:    modelID,
			Reason:   "model predates the output_config.format JSON Schema mode",
		}
	}
	// Bedrock rejects an object schema that omits additionalProperties:false;
	// surface that documented, locally-detectable requirement as a typed error.
	if err := structuredoutput.RequireClosedObjects(so.Schema); err != nil {
		return err
	}
	if input.OutputConfig == nil {
		input.OutputConfig = &anthropicOutputConfig{}
	}
	input.OutputConfig.Format = &anthropicOutputFormat{
		Type:   "json_schema",
		Schema: so.Schema,
	}
	return nil
}

// validateConverseStructuredOutput validates the Converse response against a
// per-call schema, using the shared normal-final matrix.
func validateConverseStructuredOutput(input *ConverseInput, resp *llms.ContentResponse) error {
	return validateStructuredResponse(input.StructuredOutput, input.ModelID, resp)
}

// validateStructuredResponse validates the final text of each normal-final turn
// (end_turn / stop_sequence) against the original schema. tool_use, max_tokens,
// guardrail_intervened, content_filtered and unknown reasons are intermediate or
// aborted and are not validated as final JSON. Shared by the Converse and legacy
// Anthropic paths, which use the same stop-reason vocabulary.
func validateStructuredResponse(so *llms.StructuredOutputConfig, model string, resp *llms.ContentResponse) error {
	if so == nil || resp == nil || len(resp.Choices) == 0 {
		return nil
	}
	for i, choice := range resp.Choices {
		switch choice.StopReason {
		case string(types.StopReasonEndTurn), string(types.StopReasonStopSequence):
		default:
			continue
		}
		if err := structuredoutput.Validate(so.Schema, providerBedrock, model, i, choice.StopReason, choice.Content); err != nil {
			return err
		}
	}
	return nil
}
