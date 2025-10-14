package outputparser

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vendasta/langchaingo/schema"
)

func TestBooleanParser_GetFormatInstructions(t *testing.T) {
	parser := NewBooleanParser()
	instructions := parser.GetFormatInstructions()
	assert.NotEmpty(t, instructions)
	assert.Contains(t, instructions, "`true` or `false`")
}

func TestBooleanParser_ParseWithPrompt(t *testing.T) {
	parser := NewBooleanParser()
	result, err := parser.ParseWithPrompt("yes", nil)
	require.NoError(t, err)
	boolResult, ok := result.(bool)
	require.True(t, ok)
	assert.True(t, boolResult)
}

func TestBooleanParser_Type(t *testing.T) {
	parser := NewBooleanParser()
	typ := parser.Type()
	assert.Equal(t, "boolean_parser", typ)
}

func TestCombining_GetFormatInstructions(t *testing.T) {
	parsers := []schema.OutputParser[any]{
		&Simple{},
		NewBooleanParser(),
	}
	parser := NewCombining(parsers)
	instructions := parser.GetFormatInstructions()
	assert.NotEmpty(t, instructions)
}

func TestCombining_ParseWithPrompt(t *testing.T) {
	parsers := []schema.OutputParser[any]{
		NewRegexDict(map[string]string{"key1": "Key1"}, ""),
		NewRegexDict(map[string]string{"key2": "Key2"}, ""),
	}
	parser := NewCombining(parsers)
	result, err := parser.ParseWithPrompt("Key1: value1\n\nKey2: value2", nil)
	require.NoError(t, err)
	resultMap := result.(map[string]any)
	assert.Equal(t, "value1", resultMap["key1"])
	assert.Equal(t, "value2", resultMap["key2"])
}

func TestCombining_Type(t *testing.T) {
	parser := NewCombining([]schema.OutputParser[any]{})
	typ := parser.Type()
	assert.Equal(t, "combining_parser", typ)
}

func TestCommaSeparatedList_GetFormatInstructions(t *testing.T) {
	parser := NewCommaSeparatedList()
	instructions := parser.GetFormatInstructions()
	assert.NotEmpty(t, instructions)
	assert.Contains(t, instructions, "comma separated values")
}

func TestCommaSeparatedList_ParseWithPrompt(t *testing.T) {
	parser := NewCommaSeparatedList()
	result, err := parser.ParseWithPrompt("a, b, c", nil)
	require.NoError(t, err)
	assert.Equal(t, []string{"a", "b", "c"}, result)
}

func TestCommaSeparatedList_Type(t *testing.T) {
	parser := NewCommaSeparatedList()
	typ := parser.Type()
	assert.Equal(t, "comma_separated_list_parser", typ)
}

// Test Defined parser
func TestDefined_GetFormatInstructions(t *testing.T) {
	type TestStruct struct {
		Name string `json:"name"`
	}
	parser, err := NewDefined(TestStruct{})
	require.NoError(t, err)
	instructions := parser.GetFormatInstructions()
	assert.NotEmpty(t, instructions)
	assert.Contains(t, instructions, "name")
}

func TestDefined_ParseWithPrompt(t *testing.T) {
	type TestStruct struct {
		Name string `json:"name"`
	}
	parser, err := NewDefined(TestStruct{})
	require.NoError(t, err)

	result, err := parser.ParseWithPrompt("```json\n{\"name\":\"apple\"}\n```", nil)
	require.NoError(t, err)
	// Result is the parsed struct, not an interface
	assert.Equal(t, "apple", result.Name)
}

func TestDefined_Type(t *testing.T) {
	type TestStruct struct {
		Name string `json:"name"`
	}
	parser, err := NewDefined(TestStruct{})
	require.NoError(t, err)
	typ := parser.Type()
	assert.Equal(t, "defined_parser", typ)
}

// Test RegexParser additional methods
func TestRegexParser_GetFormatInstructions(t *testing.T) {
	parser := NewRegexParser(`^\d{4}-\d{2}-\d{2}$`)
	instructions := parser.GetFormatInstructions()
	assert.NotEmpty(t, instructions)
	assert.Contains(t, instructions, "map of strings")
}

func TestRegexParser_ParseWithPrompt(t *testing.T) {
	parser := NewRegexParser(`^(\d{4})-(\d{2})-(\d{2})$`)

	result, err := parser.ParseWithPrompt("2023-12-25", nil)
	require.NoError(t, err)
	resultMap := result.(map[string]string)
	// RegexParser with capturing groups but no named groups returns map with empty string keys
	assert.Equal(t, map[string]string{"": "25"}, resultMap) // Last captured group assigned to empty key

	_, err = parser.ParseWithPrompt("invalid", nil)
	assert.Error(t, err)
}

func TestRegexParser_Type(t *testing.T) {
	parser := NewRegexParser(`test`)
	typ := parser.Type()
	assert.Equal(t, "regex_parser", typ)
}

// Test RegexDict additional methods
func TestRegexDict_GetFormatInstructions(t *testing.T) {
	parser := NewRegexDict(map[string]string{
		"name": `[A-Z][a-z]+`,
		"age":  `\d+`,
	}, "")
	instructions := parser.GetFormatInstructions()
	assert.NotEmpty(t, instructions)
}

func TestRegexDict_ParseWithPrompt(t *testing.T) {
	parser := NewRegexDict(map[string]string{
		"name": `Name`,
		"age":  `Age`,
	}, "")

	result, err := parser.ParseWithPrompt("Name: John\nAge: 30", nil)
	require.NoError(t, err)
	resultMap := result.(map[string]string)
	assert.Equal(t, "John", resultMap["name"])
	assert.Equal(t, "30", resultMap["age"])
}

func TestRegexDict_Type(t *testing.T) {
	parser := NewRegexDict(map[string]string{}, "")
	typ := parser.Type()
	assert.Equal(t, "regex_dict_parser", typ)
}

// Test Simple parser additional methods
func TestSimple_GetFormatInstructions(t *testing.T) {
	parser := &Simple{}
	instructions := parser.GetFormatInstructions()
	assert.Empty(t, instructions)
}

func TestSimple_ParseWithPrompt(t *testing.T) {
	parser := &Simple{}
	result, err := parser.ParseWithPrompt("test", nil)
	require.NoError(t, err)
	assert.Equal(t, "test", result)
}

func TestSimple_Type(t *testing.T) {
	parser := &Simple{}
	typ := parser.Type()
	assert.Equal(t, "simple_parser", typ)
}

// Test Structured parser additional methods
func TestStructured_GetFormatInstructions(t *testing.T) {
	schemas := []ResponseSchema{
		{Name: "name", Description: "person's name"},
		{Name: "age", Description: "person's age"},
	}

	parser := NewStructured(schemas)
	instructions := parser.GetFormatInstructions()
	assert.NotEmpty(t, instructions)
	assert.Contains(t, instructions, "json")
}

func TestStructured_ParseWithPrompt(t *testing.T) {
	schemas := []ResponseSchema{
		{Name: "name", Description: "person's name"},
		{Name: "age", Description: "person's age"},
	}

	parser := NewStructured(schemas)
	result, err := parser.ParseWithPrompt("```json\n{\"name\":\"John\",\"age\":\"30\"}\n```", nil)
	require.NoError(t, err)

	resultMap := result.(map[string]string)
	assert.Equal(t, "John", resultMap["name"])
	assert.Equal(t, "30", resultMap["age"])
}

func TestStructured_Type(t *testing.T) {
	parser := NewStructured([]ResponseSchema{})
	typ := parser.Type()
	assert.Equal(t, "structured_parser", typ)
}

// Test error conditions
func TestParsers_ErrorHandling(t *testing.T) {
	tests := []struct {
		name    string
		parser  schema.OutputParser[any]
		input   string
		wantErr bool
	}{
		{
			name:    "BooleanParser invalid input",
			parser:  NewBooleanParser(),
			input:   "maybe",
			wantErr: true,
		},
		{
			name:    "RegexParser no match",
			parser:  NewRegexParser(`^\d+$`),
			input:   "abc",
			wantErr: true,
		},
		{
			name:    "Structured parser invalid JSON",
			parser:  NewStructured([]ResponseSchema{{Name: "test"}}),
			input:   "not json",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.parser.Parse(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Test combining parser with different parsers
func TestCombining_WithDifferentParsers(t *testing.T) {
	// Create parsers that return map[string]string for combining parser
	regexParser1 := NewRegexDict(map[string]string{"name": "Name"}, "")
	regexParser2 := NewRegexDict(map[string]string{"age": "Age"}, "")

	combiner := NewCombining([]schema.OutputParser[any]{
		regexParser1,
		regexParser2,
	})

	// Test parsing with multiple text chunks separated by double newlines
	// Format expected by RegexDict: "Key: value"
	text := "Name: John\n\nAge: 30"
	result, err := combiner.Parse(text)
	require.NoError(t, err)

	resultMap := result.(map[string]any)
	assert.NotEmpty(t, resultMap)
}
