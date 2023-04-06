package outputParsers_test

import (
	"reflect"
	"testing"

	"github.com/tmc/langchaingo/outputParsers"
)

func TestStructuredOutputParserFromNameAndDescription(t *testing.T) {
	parser := outputParsers.NewStructuredFromNameAndDescription(map[string]string{"url": "A link to the resource"})

	instruction := parser.GetFormatInstructions()
	expectedInstruction := "The output should be a markdown code snippet formatted in the following schema: \n```json\n{\n\t\"url\": string // A link to the resource\n}\n```"
	if instruction != expectedInstruction {
		t.Errorf("StructuredOutputParser format instruction is not equal expected. \n Got:\n %s Expected: \n %s", instruction, expectedInstruction)
	}

	llmOutput := "```json\n{\n\t\"url\": \"https://google.com\" \n}\n```"

	parsed, err := parser.Parse(llmOutput)
	if err != nil {
		t.Errorf("Unexpected error parsing: %e", err)
	}

	expectedParsed := map[string]string{"url": "https://google.com"}

	if !reflect.DeepEqual(parsed, expectedParsed) {
		t.Errorf("StructuredOutputParser parsing result not equal expected: Got: %v. Expected: %v", parsed, expectedParsed)
	}
}
