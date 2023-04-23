package output_parser

import (
	"fmt"
	"strings"

	"github.com/tmc/langchaingo/schema"
)

const (
	_structuredFormatInstructionTemplate = "The output should be a markdown code snippet formatted in the following schema: \n```json\n{\n%s}\n```"
	_structuredLineTemplate              = "%s: %s // %s"
)

type ResponseSchema struct {
	Name        string
	Description string
}

type Structured struct {
	ResponseSchemas []ResponseSchema
}

func NewStructured(schema []ResponseSchema) Structured {
	return Structured{
		ResponseSchemas: schema,
	}
}

var _ schema.OutputParser[map[string]string] = Structured{}

func (p Structured) Parse(text string) (map[string]string, error) {
	withoutJSONStart := strings.Split(text, "```json")
	if len(withoutJSONStart) < 2 {
		return nil, fmt.Errorf("Failed to parse. Text: %s. Error: no ```json at start of output", text)
	}

	withoutJSONEnd := strings.Split(withoutJSONStart[1], "```")
	if len(withoutJSONEnd) < 1 {
		return nil, fmt.Errorf("Failed to parse. Text: %s. Error: no ```json at end of output", text)
	}

	jsonString := withoutJSONEnd[0]

}

func (p Structured) ParseWithPrompt(text string, prompt schema.PromptValue) (map[string]string, error) {
	return p.Parse(text)
}

/* // ParseWithPrompt parses the output of an LLM call with the prompt used.
   ParseWithPrompt(text string, prompt PromptValue) (T, error)
   // GetFormatInstructions returns a string describing the format of the output.
   GetFormatInstructions() string
   // Type returns the string type key uniquely identifying this class of parser
   Type() string */

func (p Structured) GetFormatInstructions() string {
	formatInstructionJSONLines := ""

	return fmt.Sprintf(_structuredFormatInstructionTemplate, formatInstructionJSONLines)
}

func (p Structured) Type() string {
	return "structured_parser"
}
