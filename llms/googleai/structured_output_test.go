package googleai

import (
	"testing"

	"github.com/google/generative-ai-go/genai"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/llms"
)

func TestConvertToResponseSchema(t *testing.T) {
	tests := []struct {
		name    string
		input   *llms.StructuredOutputSchema
		want    *genai.Schema
		wantErr bool
	}{
		{
			name: "simple string",
			input: &llms.StructuredOutputSchema{
				Type:        llms.SchemaTypeString,
				Description: "A string value",
			},
			want: &genai.Schema{
				Type:        genai.TypeString,
				Description: "A string value",
			},
			wantErr: false,
		},
		{
			name: "simple object",
			input: &llms.StructuredOutputSchema{
				Type:        llms.SchemaTypeObject,
				Description: "A user object",
				Properties: map[string]*llms.StructuredOutputSchema{
					"name": {
						Type:        llms.SchemaTypeString,
						Description: "User's name",
					},
					"age": {
						Type:        llms.SchemaTypeInteger,
						Description: "User's age",
					},
				},
				Required: []string{"name", "age"},
			},
			want: &genai.Schema{
				Type:        genai.TypeObject,
				Description: "A user object",
				Properties: map[string]*genai.Schema{
					"name": {
						Type:        genai.TypeString,
						Description: "User's name",
					},
					"age": {
						Type:        genai.TypeInteger,
						Description: "User's age",
					},
				},
				Required: []string{"name", "age"},
			},
			wantErr: false,
		},
		{
			name: "nested object",
			input: &llms.StructuredOutputSchema{
				Type: llms.SchemaTypeObject,
				Properties: map[string]*llms.StructuredOutputSchema{
					"user": {
						Type: llms.SchemaTypeObject,
						Properties: map[string]*llms.StructuredOutputSchema{
							"name":  {Type: llms.SchemaTypeString},
							"email": {Type: llms.SchemaTypeString},
						},
						Required: []string{"name"},
					},
				},
			},
			want: &genai.Schema{
				Type: genai.TypeObject,
				Properties: map[string]*genai.Schema{
					"user": {
						Type: genai.TypeObject,
						Properties: map[string]*genai.Schema{
							"name":  {Type: genai.TypeString},
							"email": {Type: genai.TypeString},
						},
						Required: []string{"name"},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "array of strings",
			input: &llms.StructuredOutputSchema{
				Type:        llms.SchemaTypeArray,
				Description: "List of tags",
				Items: &llms.StructuredOutputSchema{
					Type: llms.SchemaTypeString,
				},
			},
			want: &genai.Schema{
				Type:        genai.TypeArray,
				Description: "List of tags",
				Items: &genai.Schema{
					Type: genai.TypeString,
				},
			},
			wantErr: false,
		},
		{
			name: "array of objects",
			input: &llms.StructuredOutputSchema{
				Type: llms.SchemaTypeArray,
				Items: &llms.StructuredOutputSchema{
					Type: llms.SchemaTypeObject,
					Properties: map[string]*llms.StructuredOutputSchema{
						"id":   {Type: llms.SchemaTypeInteger},
						"name": {Type: llms.SchemaTypeString},
					},
					Required: []string{"id"},
				},
			},
			want: &genai.Schema{
				Type: genai.TypeArray,
				Items: &genai.Schema{
					Type: genai.TypeObject,
					Properties: map[string]*genai.Schema{
						"id":   {Type: genai.TypeInteger},
						"name": {Type: genai.TypeString},
					},
					Required: []string{"id"},
				},
			},
			wantErr: false,
		},
		{
			name: "string with enum",
			input: &llms.StructuredOutputSchema{
				Type:        llms.SchemaTypeString,
				Description: "Color choice",
				Enum:        []string{"red", "green", "blue"},
			},
			want: &genai.Schema{
				Type:        genai.TypeString,
				Description: "Color choice",
				Enum:        []string{"red", "green", "blue"},
			},
			wantErr: false,
		},
		{
			name: "all primitive types",
			input: &llms.StructuredOutputSchema{
				Type: llms.SchemaTypeObject,
				Properties: map[string]*llms.StructuredOutputSchema{
					"string_field":  {Type: llms.SchemaTypeString},
					"integer_field": {Type: llms.SchemaTypeInteger},
					"number_field":  {Type: llms.SchemaTypeNumber},
					"boolean_field": {Type: llms.SchemaTypeBoolean},
					"null_field":    {Type: llms.SchemaTypeNull},
				},
			},
			want: &genai.Schema{
				Type: genai.TypeObject,
				Properties: map[string]*genai.Schema{
					"string_field":  {Type: genai.TypeString},
					"integer_field": {Type: genai.TypeInteger},
					"number_field":  {Type: genai.TypeNumber},
					"boolean_field": {Type: genai.TypeBoolean},
					"null_field":    {Type: genai.TypeUnspecified},
				},
			},
			wantErr: false,
		},
		{
			name: "complex recipe schema",
			input: &llms.StructuredOutputSchema{
				Type: llms.SchemaTypeObject,
				Properties: map[string]*llms.StructuredOutputSchema{
					"title": {
						Type:        llms.SchemaTypeString,
						Description: "Recipe title",
					},
					"servings": {
						Type:        llms.SchemaTypeInteger,
						Description: "Number of servings",
					},
					"ingredients": {
						Type: llms.SchemaTypeArray,
						Items: &llms.StructuredOutputSchema{
							Type: llms.SchemaTypeObject,
							Properties: map[string]*llms.StructuredOutputSchema{
								"name": {
									Type:        llms.SchemaTypeString,
									Description: "Ingredient name",
								},
								"amount": {
									Type:        llms.SchemaTypeString,
									Description: "Amount needed",
								},
							},
							Required: []string{"name", "amount"},
						},
					},
					"steps": {
						Type: llms.SchemaTypeArray,
						Items: &llms.StructuredOutputSchema{
							Type: llms.SchemaTypeString,
						},
					},
				},
				Required: []string{"title", "ingredients", "steps"},
			},
			want: &genai.Schema{
				Type: genai.TypeObject,
				Properties: map[string]*genai.Schema{
					"title": {
						Type:        genai.TypeString,
						Description: "Recipe title",
					},
					"servings": {
						Type:        genai.TypeInteger,
						Description: "Number of servings",
					},
					"ingredients": {
						Type: genai.TypeArray,
						Items: &genai.Schema{
							Type: genai.TypeObject,
							Properties: map[string]*genai.Schema{
								"name": {
									Type:        genai.TypeString,
									Description: "Ingredient name",
								},
								"amount": {
									Type:        genai.TypeString,
									Description: "Amount needed",
								},
							},
							Required: []string{"name", "amount"},
						},
					},
					"steps": {
						Type: genai.TypeArray,
						Items: &genai.Schema{
							Type: genai.TypeString,
						},
					},
				},
				Required: []string{"title", "ingredients", "steps"},
			},
			wantErr: false,
		},
		{
			name:    "nil schema",
			input:   nil,
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := convertToResponseSchema(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestConvertSchemaType(t *testing.T) {
	tests := []struct {
		input llms.SchemaType
		want  genai.Type
	}{
		{llms.SchemaTypeObject, genai.TypeObject},
		{llms.SchemaTypeArray, genai.TypeArray},
		{llms.SchemaTypeString, genai.TypeString},
		{llms.SchemaTypeNumber, genai.TypeNumber},
		{llms.SchemaTypeInteger, genai.TypeInteger},
		{llms.SchemaTypeBoolean, genai.TypeBoolean},
		{llms.SchemaTypeNull, genai.TypeUnspecified},
		{llms.SchemaType("unknown"), genai.TypeUnspecified},
	}

	for _, tt := range tests {
		t.Run(string(tt.input), func(t *testing.T) {
			got := convertSchemaType(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}
