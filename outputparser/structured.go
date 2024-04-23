package outputparser

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/schema"
)

// ParseError is the error type returned by output parsers.
type ParseError struct {
	Text   string
	Reason string
}

func (e ParseError) Error() string {
	return fmt.Sprintf("parse text %s. %s", e.Text, e.Reason)
}

const (
	// _structuredFormatInstructionTemplate is a template for the format
	// instructions of the structured output parser.
	_structuredFormatInstructionTemplate = "The output should be a markdown code snippet formatted in the following schema: \n```json\n{\n%s}\n```" // nolint

	// _structuredLineTemplate is a single line of the json schema in the
	// format instruction of the structured output parser. The fist verb is
	// the name, the second verb is the type and the third is a description of
	// what the field should contain.
	_structuredLineTemplate = "\"%s\": %s // %s\n"
)

// ResponseSchema is struct used in the structured output parser to describe
// how the llm should format its response. Name is a key in the parsed
// output map. Description is a description of what the value should contain.
type ResponseSchema struct {
	Name        string
	Description string
}

// Structured is an output parser that parses the output of an LLM into key value
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

// Statically assert that Structured implement the OutputParser interface.
var _ schema.OutputParser[any] = Structured{}

// Parse parses the output of an LLM into a map. If the output of the llm doesn't
// contain every filed specified in the response schemas, the function will return
// an error.
func (p Structured) parse(text string) (map[string]string, error) {
	// Remove the ```json that should be at the start of the text, and the ```
	// that should be at the end of the text.
	withoutJSONStart := strings.Split(text, "```json")
	if !(len(withoutJSONStart) > 1) {
		return nil, ParseError{Text: text, Reason: "no ```json at start of output"}
	}

	withoutJSONEnd := strings.Split(withoutJSONStart[1], "```")
	if len(withoutJSONEnd) < 1 {
		return nil, ParseError{Text: text, Reason: "no ``` at end of output"}
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
		return nil, ParseError{
			Text:   text,
			Reason: fmt.Sprintf("output is missing the following fields %v", missingKeys),
		}
	}

	return parsed, nil
}

func (p Structured) Parse(text string) (any, error) {
	return p.parse(text)
}

// ParseWithPrompt does the same as Parse.
func (p Structured) ParseWithPrompt(text string, _ llms.PromptValue) (any, error) {
	return p.parse(text)
}

// GetFormatInstructions returns a string explaining how the llm should format
// its response.
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
