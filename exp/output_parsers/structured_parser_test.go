package output_parsers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStructuredOutputParserFromNameAndDescription(t *testing.T) {
	t.Parallel()

	parser := NewStructuredFromNameAndDescription(map[string]string{"url": "A link to the resource"})

	instruction := parser.GetFormatInstructions()
	expectedInstruction := "" +
		"The output should be a markdown code snippet formatted in the following schema" +
		": \n```json\n{\n\t\"url\": string // A link to the resource\n}\n```"
	assert.Equal(t, expectedInstruction, instruction)

	llmOutput := "```json\n{\n\t\"url\": \"https://google.com\" \n}\n```"

	parsed, err := parser.Parse(llmOutput)
	require.NoError(t, err)

	expectedParsed := map[string]string{"url": "https://google.com"}
	assert.Equal(t, expectedParsed, parsed)
}
