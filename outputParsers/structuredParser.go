package outputParsers

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/tmc/langchaingo/prompts"
)

type StructuredOutputParser struct {
	schema string
}

func NewStructuredFromNameAndDescription(schemaValues map[string]string) StructuredOutputParser {
	s := StructuredOutputParser{
		schema: "",
	}

	keyValueJSONString := ""
	for name, description := range schemaValues {
		keyValueJSONString += fmt.Sprintf("\t\"%s\": string // %s\n", name, description)
	}

	s.schema = "{\n" + keyValueJSONString + "}\n"

	return s
}

func (p StructuredOutputParser) GetFormatInstructions() string {
	return "The output should be a markdown code snippet formatted in the following schema: \n```json\n" + p.schema + "```"
}

func (p StructuredOutputParser) Parse(text string) (any, error) {
	withoutJSONStart := strings.Split(text, "```json")
	if len(withoutJSONStart) < 2 {
		return map[string]string{}, OutputParserException{Reason: fmt.Sprintf("Failed to parse. Text: %s. Error: no ```json at start of output", text)}
	}

	withoutJSONEnd := strings.Split(withoutJSONStart[1], "```")
	if len(withoutJSONEnd) < 1 {
		return map[string]string{}, OutputParserException{Reason: fmt.Sprintf("Failed to parse. Text: %s. Error: no ```json at ebd of output", text)}
	}

	jsonString := withoutJSONEnd[0]

	var parsed map[string]string
	err := json.Unmarshal([]byte(jsonString), &parsed)
	if err != nil {
		return parsed, OutputParserException{Reason: fmt.Sprintf("Failed to parse. Text: %s. Error: %e", text, err)}
	}

	return parsed, nil
}

func (p StructuredOutputParser) ParseWithPrompt(text string, prompt prompts.PromptValue) (any, error) {
	return p.Parse(text)
}
