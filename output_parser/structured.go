package output_parser

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/tmc/langchaingo/schema"
)

const (
	// _structuredFormatInstructionTemplate is a template for the format
	// instructions of the structured output parser.
	_structuredFormatInstructionTemplate = "The output should be a markdown code snippet formatted in the following schema: \n```json\n{\n%s}\n```"

	// _structuredLineTemplate is a single line of the json schema in the
	// format instruction of the structured output parser. The fist verb is
	// the name, the second verb is the type and the third is a description of
	// what the field should contain.
	_structuredLineTemplate = "\"%s\": %s // %s\n"
)

// ResponseSchema is struct used in the structured output parser to describe
// how the llm should format it's response. The name field of the struct is a
// key in the parsed output map. The description
type ResponseSchema struct {
	Name        string
	Description string
}

// Structured is an output parser that parses the output of an llm into key value
// pairs. The name and description of what values the output of the llm should
// contain is stored in a list of response schema.
type Structured struct {
	ResponseSchemas []ResponseSchema
}

// NewStructured is a function that creates a new structured output parser from
// a list of response schemas.
func NewStructured(schema []ResponseSchema) Structured {
	return Structured{
		ResponseSchemas: schema,
	}
}

// Statically assert that CommaSeparatedList implement the OutputParser interface.
var _ schema.OutputParser[map[string]string] = Structured{}

// Parse parses the output of an llm into a map. If the output of the llm doesn't
// contain every filed specified in the response schemas, the function will return
// an error.
func (p Structured) Parse(text string) (map[string]string, error) {
	// Remove the ```json that should be at the start of the text, and the ```
	// that should be at the end of the text.
	withoutJSONStart := strings.Split(text, "```json")
	if len(withoutJSONStart) < 2 {
		return nil, fmt.Errorf("Text: %s. Error: no ```json at start of output", text)
	}

	withoutJSONEnd := strings.Split(withoutJSONStart[1], "```")
	if len(withoutJSONEnd) < 1 {
		return nil, fmt.Errorf("Text: %s. Error: no ``` at end of output", text)
	}

	jsonString := withoutJSONEnd[0]

	var parsed map[string]string
	err := json.Unmarshal([]byte(jsonString), &parsed)
	if err != nil {
		return nil, err
	}

	// Validate that the parsed map contains all fields specified in the response
	// schemas.
	missingKeys := make([]string, 0)
	for _, rs := range p.ResponseSchemas {
		if _, ok := parsed[rs.Name]; !ok {
			missingKeys = append(missingKeys, rs.Name)
		}
	}

	if len(missingKeys) > 0 {
		return nil, fmt.Errorf("Text: %s. Error: output is missing the following fields %v", text, missingKeys)
	}

	return parsed, nil
}

// ParseWithPrompt does the same as Parse.
func (p Structured) ParseWithPrompt(text string, prompt schema.PromptValue) (map[string]string, error) {
	return p.Parse(text)
}

// GetFormatInstructions returns a string explaining how the llm should format
// it's response.
func (p Structured) GetFormatInstructions() string {
	jsonLines := ""
	for _, rs := range p.ResponseSchemas {
		jsonLines += "\t" + fmt.Sprintf(
			_structuredLineTemplate,
			rs.Name,
			"string", /* type of the filed*/
			rs.Description,
		)
	}

	return fmt.Sprintf(_structuredFormatInstructionTemplate, jsonLines)
}

// Type returns the type of the output parser.
func (p Structured) Type() string {
	return "structured_parser"
}
