// This file is hand-written (not generated). It builds Vertex's response schema
// from the provider-neutral llms.StructuredOutput. The Vertex SDK accepts only a
// *genai.Schema (a subset of OpenAPI 3.0), so the raw JSON Schema is converted
// explicitly; any construct that cannot be represented without changing meaning is
// a typed error rather than a silent drop.
package vertex

import (
	"bytes"
	"encoding/json"
	"fmt"

	"cloud.google.com/go/vertexai/genai"

	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/llms/structuredoutput"
)

const providerVertex = "vertex"

// applyVertexResponseFormat sets the model's response format from the call options.
// With a per-call StructuredOutput it sets ResponseSchema (converted from the raw
// JSON Schema) plus application/json; otherwise it keeps the pre-existing schema-
// less JSON mode / ResponseMIMEType behavior.
func applyVertexResponseFormat(model *genai.GenerativeModel, opts *llms.CallOptions) error {
	if so := opts.StructuredOutput; so != nil {
		if err := opts.ValidateStructuredOutput(); err != nil {
			return err
		}
		switch opts.GetResponseMIMEType() {
		case "", ResponseMIMETypeJson:
			model.ResponseMIMEType = ResponseMIMETypeJson
		default:
			return &llms.ErrStructuredOutputConflict{
				Provider: providerVertex,
				Detail:   "structured output requires the application/json response MIME type",
			}
		}
		schema, err := convertJSONSchemaToVertex(so.Schema)
		if err != nil {
			return err
		}
		model.ResponseSchema = schema
		return nil
	}

	switch {
	case opts.ResponseMIMEType != nil && opts.JSONMode:
		return fmt.Errorf("conflicting options, can't use JSONMode and ResponseMIMEType together")
	case opts.ResponseMIMEType != nil && !opts.JSONMode:
		model.ResponseMIMEType = opts.GetResponseMIMEType()
	case opts.GetResponseMIMEType() == "" && opts.JSONMode:
		model.ResponseMIMEType = ResponseMIMETypeJson
	}
	return nil
}

// validateVertexStructuredOutput validates each normal-final (FinishReasonStop)
// candidate against the original schema. Other finish reasons keep their prior
// semantics and are not validated as final JSON.
func validateVertexStructuredOutput(opts *llms.CallOptions, resp *llms.ContentResponse) error {
	so := opts.StructuredOutput
	if so == nil || resp == nil {
		return nil
	}
	model := opts.GetModel()
	stop := genai.FinishReasonStop.String()
	for i, choice := range resp.Choices {
		if choice.StopReason != stop {
			continue
		}
		if err := structuredoutput.Validate(so.Schema, providerVertex, model, i, choice.StopReason, choice.Content); err != nil {
			return err
		}
	}
	return nil
}

func unsupportedVertexSchema(reason string) error {
	return &llms.ErrStructuredOutputUnsupported{Provider: providerVertex, Reason: reason}
}

// vertexUnrepresentable lists JSON Schema keywords that *genai.Schema cannot
// express; their presence is a hard error so a constraint is never silently lost.
var vertexUnrepresentable = []string{
	"$ref", "$defs", "definitions", "anyOf", "oneOf", "allOf", "not",
	"if", "then", "else", "const", "patternProperties", "dependentSchemas",
}

// vertexKnownKeywords are the keywords convertVertexNode understands: the ones it
// maps onto genai.Schema plus purely annotative metadata that is safe to ignore.
// Anything outside this set is rejected rather than silently dropped.
var vertexKnownKeywords = map[string]bool{
	// mapped onto genai.Schema
	"type": true, "description": true, "title": true, "format": true,
	"nullable": true, "enum": true, "items": true, "properties": true,
	"required": true, "additionalProperties": true,
	"minimum": true, "maximum": true, "minItems": true, "maxItems": true,
	"minProperties": true, "maxProperties": true,
	"minLength": true, "maxLength": true, "pattern": true,
	// annotative only — no effect on validation, safe to ignore
	"$schema": true, "$id": true, "$comment": true, "default": true,
	"examples": true, "readOnly": true, "writeOnly": true, "deprecated": true,
}

// convertJSONSchemaToVertex converts a raw JSON Schema document to *genai.Schema.
// Numbers decode via json.Number so integer keywords can be checked for
// integrality instead of being silently truncated.
func convertJSONSchemaToVertex(raw json.RawMessage) (*genai.Schema, error) {
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	var node map[string]any
	if err := dec.Decode(&node); err != nil {
		return nil, unsupportedVertexSchema(fmt.Sprintf("schema is not a JSON object: %v", err))
	}
	return convertVertexNode(node)
}

func convertVertexNode(node map[string]any) (*genai.Schema, error) { //nolint:funlen // linear JSON Schema -> *genai.Schema field mapping
	for _, k := range vertexUnrepresentable {
		if _, ok := node[k]; ok {
			return nil, unsupportedVertexSchema(fmt.Sprintf("construct %q is not representable by the Vertex schema type", k))
		}
	}
	// Reject any keyword we neither map nor knowingly ignore, so a real constraint
	// (multipleOf, exclusiveMinimum, uniqueItems, ...) is never silently lost.
	for k := range node {
		if !vertexKnownKeywords[k] {
			return nil, unsupportedVertexSchema(fmt.Sprintf("keyword %q is not representable by the Vertex schema type", k))
		}
	}

	schema := &genai.Schema{}

	gt, nullable, err := vertexTypeOf(node)
	if err != nil {
		return nil, err
	}
	schema.Type = gt
	schema.Nullable = nullable

	if s, ok := node["description"].(string); ok {
		schema.Description = s
	}
	if s, ok := node["title"].(string); ok {
		schema.Title = s
	}
	if s, ok := node["format"].(string); ok {
		schema.Format = s
	}
	if s, ok := node["pattern"].(string); ok {
		schema.Pattern = s
	}

	// additionalProperties: only the strict `false` is accepted. Vertex object
	// schemas have no field for it, but dropping `false` does not loosen the
	// effective guarantee — the final response is validated against the original
	// schema locally. Any schema value (or `true`) IS a meaningful loosening, so it
	// is rejected instead of silently ignored.
	if ap, ok := node["additionalProperties"]; ok {
		if b, isBool := ap.(bool); !isBool || b {
			return nil, unsupportedVertexSchema("additionalProperties must be false; a schema or true is not representable")
		}
	}

	if enum, ok := node["enum"]; ok {
		arr, ok := enum.([]any)
		if !ok {
			return nil, unsupportedVertexSchema("enum must be an array")
		}
		for _, v := range arr {
			s, ok := v.(string)
			if !ok {
				return nil, unsupportedVertexSchema("Vertex enum supports only string values")
			}
			schema.Enum = append(schema.Enum, s)
		}
	}

	if items, ok := node["items"]; ok {
		itemMap, ok := items.(map[string]any)
		if !ok {
			return nil, unsupportedVertexSchema("items must be a single schema object (tuple items are not representable)")
		}
		child, err := convertVertexNode(itemMap)
		if err != nil {
			return nil, err
		}
		schema.Items = child
	}

	if props, ok := node["properties"]; ok {
		propMap, ok := props.(map[string]any)
		if !ok {
			return nil, unsupportedVertexSchema("properties must be an object")
		}
		schema.Properties = make(map[string]*genai.Schema, len(propMap))
		for name, raw := range propMap {
			childMap, ok := raw.(map[string]any)
			if !ok {
				return nil, unsupportedVertexSchema(fmt.Sprintf("property %q must be a schema object", name))
			}
			child, err := convertVertexNode(childMap)
			if err != nil {
				return nil, err
			}
			schema.Properties[name] = child
		}
	}

	if req, hasReq := node["required"]; hasReq {
		arr, ok := req.([]any)
		if !ok {
			return nil, unsupportedVertexSchema("required must be an array of strings")
		}
		for _, r := range arr {
			s, ok := r.(string)
			if !ok {
				return nil, unsupportedVertexSchema("required must contain only strings")
			}
			schema.Required = append(schema.Required, s)
		}
	}

	return schema, setVertexNumeric(schema, node)
}

// vertexTypeOf resolves the genai type from a "type" that is a string or a
// [T, "null"] pair (the common JSON Schema nullable idiom). A wider union is not
// representable.
func vertexTypeOf(node map[string]any) (genai.Type, bool, error) {
	rawType, ok := node["type"]
	if !ok {
		// Infer from shape when the type keyword is absent.
		if _, hasProps := node["properties"]; hasProps {
			return genai.TypeObject, false, nil
		}
		if _, hasItems := node["items"]; hasItems {
			return genai.TypeArray, false, nil
		}
		return genai.TypeUnspecified, false, unsupportedVertexSchema("schema node is missing a type")
	}
	switch tv := rawType.(type) {
	case string:
		gt, err := mapVertexType(tv)
		return gt, false, err
	case []any:
		var nonNull string
		count, nullable := 0, false
		for _, e := range tv {
			s, _ := e.(string)
			if s == "null" {
				nullable = true
				continue
			}
			nonNull = s
			count++
		}
		if count != 1 {
			return genai.TypeUnspecified, false, unsupportedVertexSchema("union types are not representable (only [T, \"null\"] is)")
		}
		gt, err := mapVertexType(nonNull)
		return gt, nullable, err
	default:
		return genai.TypeUnspecified, false, unsupportedVertexSchema("type must be a string or array")
	}
}

func mapVertexType(t string) (genai.Type, error) {
	switch t {
	case "string":
		return genai.TypeString, nil
	case "number":
		return genai.TypeNumber, nil
	case "integer":
		return genai.TypeInteger, nil
	case "boolean":
		return genai.TypeBoolean, nil
	case "array":
		return genai.TypeArray, nil
	case "object":
		return genai.TypeObject, nil
	default:
		return genai.TypeUnspecified, unsupportedVertexSchema(fmt.Sprintf("unsupported type %q", t))
	}
}

// setVertexNumeric transfers numeric keywords. Values arrive as json.Number
// (UseNumber). Float-valued keywords parse as float64; integer-valued keywords
// (item/property/length counts) must be integral — a fractional value is a typed
// error rather than a silent truncation.
func setVertexNumeric(schema *genai.Schema, node map[string]any) error {
	floats := map[string]*float64{"minimum": &schema.Minimum, "maximum": &schema.Maximum}
	for key, dst := range floats {
		n, ok := node[key].(json.Number)
		if !ok {
			continue
		}
		f, err := n.Float64()
		if err != nil {
			return unsupportedVertexSchema(fmt.Sprintf("%s must be a number", key))
		}
		*dst = f
	}
	ints := map[string]*int64{
		"minItems": &schema.MinItems, "maxItems": &schema.MaxItems,
		"minProperties": &schema.MinProperties, "maxProperties": &schema.MaxProperties,
		"minLength": &schema.MinLength, "maxLength": &schema.MaxLength,
	}
	for key, dst := range ints {
		n, ok := node[key].(json.Number)
		if !ok {
			continue
		}
		i, err := n.Int64()
		if err != nil {
			return unsupportedVertexSchema(fmt.Sprintf("%s must be an integer", key))
		}
		*dst = i
	}
	return nil
}
