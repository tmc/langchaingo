package llms

// StructuredOutputDefinition defines the schema for structured output responses.
// This provides a unified API across providers (OpenAI, Anthropic, Google) for
// requesting JSON-structured responses that conform to a specific schema.
//
// Example usage:
//
//	schema := &llms.StructuredOutputDefinition{
//	    Name:        "user_info",
//	    Description: "Extract user information",
//	    Schema: &llms.StructuredOutputSchema{
//	        Type: llms.SchemaTypeObject,
//	        Properties: map[string]*llms.StructuredOutputSchema{
//	            "name": {Type: llms.SchemaTypeString, Description: "User's full name"},
//	            "age":  {Type: llms.SchemaTypeInteger, Description: "User's age in years"},
//	        },
//	        Required: []string{"name", "age"},
//	    },
//	}
//
//	resp, err := llm.GenerateContent(ctx, messages,
//	    llms.WithStructuredOutput(schema),
//	)
type StructuredOutputDefinition struct {
	// Name is the schema identifier
	Name string

	// Description explains what this schema represents
	Description string

	// Schema is the actual JSON schema definition
	Schema *StructuredOutputSchema

	// Strict enforces exact schema matching (OpenAI only).
	// When true, the response must match the schema exactly with no additional fields.
	// For other providers, this is advisory and may not be enforced.
	Strict bool
}

// StructuredOutputSchema defines a JSON schema for structured output.
// This schema format is compatible with JSON Schema and can be translated
// to provider-specific formats (OpenAI ResponseFormat, Google ResponseSchema, etc.).
type StructuredOutputSchema struct {
	// Type specifies the data type (object, array, string, number, integer, boolean)
	Type SchemaType

	// Properties defines the properties of an object (for SchemaTypeObject)
	Properties map[string]*StructuredOutputSchema

	// Items defines the schema for array items (for SchemaTypeArray)
	Items *StructuredOutputSchema

	// Required lists required property names (for SchemaTypeObject)
	Required []string

	// AdditionalProperties controls whether extra properties are allowed (for SchemaTypeObject)
	AdditionalProperties bool

	// Description provides documentation for this schema element
	Description string

	// Enum restricts values to a specific set (for string/number types)
	Enum []string

	// Minimum sets the minimum value (for number/integer types)
	Minimum *float64

	// Maximum sets the maximum value (for number/integer types)
	Maximum *float64

	// MinLength sets minimum string length (for string type)
	MinLength *int

	// MaxLength sets maximum string length (for string type)
	MaxLength *int

	// Pattern is a regex pattern for validation (for string type)
	Pattern string

	// MinItems sets minimum array length (for array type)
	MinItems *int

	// MaxItems sets maximum array length (for array type)
	MaxItems *int
}

// SchemaType represents JSON schema types
type SchemaType string

const (
	// SchemaTypeObject represents a JSON object with properties
	SchemaTypeObject SchemaType = "object"

	// SchemaTypeArray represents a JSON array
	SchemaTypeArray SchemaType = "array"

	// SchemaTypeString represents a string value
	SchemaTypeString SchemaType = "string"

	// SchemaTypeNumber represents a floating-point number
	SchemaTypeNumber SchemaType = "number"

	// SchemaTypeInteger represents an integer
	SchemaTypeInteger SchemaType = "integer"

	// SchemaTypeBoolean represents a boolean value
	SchemaTypeBoolean SchemaType = "boolean"

	// SchemaTypeNull represents a null value
	SchemaTypeNull SchemaType = "null"
)

// WithStructuredOutput configures the LLM to return JSON output conforming to the given schema.
//
// Provider Support:
//   - OpenAI: Native support via ResponseFormat with json_schema type
//   - Google: Native support via ResponseSchema
//   - Anthropic: Simulated via tool calling (the schema is converted to a tool definition)
//
// Example:
//
//	schema := &llms.StructuredOutputDefinition{
//	    Name: "recipe",
//	    Schema: &llms.StructuredOutputSchema{
//	        Type: llms.SchemaTypeObject,
//	        Properties: map[string]*llms.StructuredOutputSchema{
//	            "title":       {Type: llms.SchemaTypeString},
//	            "ingredients": {Type: llms.SchemaTypeArray, Items: &llms.StructuredOutputSchema{Type: llms.SchemaTypeString}},
//	            "steps":       {Type: llms.SchemaTypeArray, Items: &llms.StructuredOutputSchema{Type: llms.SchemaTypeString}},
//	        },
//	        Required: []string{"title", "ingredients", "steps"},
//	    },
//	}
//
//	resp, err := llm.GenerateContent(ctx, messages, llms.WithStructuredOutput(schema))
func WithStructuredOutput(schema *StructuredOutputDefinition) CallOption {
	return func(o *CallOptions) {
		o.StructuredOutput = schema
	}
}

// ValidateSchema performs basic validation on a StructuredOutputSchema.
// Returns an error if the schema is invalid.
func ValidateSchema(schema *StructuredOutputSchema) error {
	if schema == nil {
		return ErrInvalidSchema
	}

	switch schema.Type {
	case SchemaTypeObject:
		if len(schema.Properties) == 0 {
			return ErrInvalidSchema
		}
		// Recursively validate nested schemas
		for _, prop := range schema.Properties {
			if err := ValidateSchema(prop); err != nil {
				return err
			}
		}

	case SchemaTypeArray:
		if schema.Items == nil {
			return ErrInvalidSchema
		}
		// Recursively validate items schema
		if err := ValidateSchema(schema.Items); err != nil {
			return err
		}

	case SchemaTypeString, SchemaTypeNumber, SchemaTypeInteger, SchemaTypeBoolean, SchemaTypeNull:
		// Primitive types are always valid

	default:
		return ErrInvalidSchema
	}

	return nil
}

// ErrInvalidSchema is returned when a schema is invalid
var ErrInvalidSchema = NewError("invalid schema", "the provided schema is not valid")
