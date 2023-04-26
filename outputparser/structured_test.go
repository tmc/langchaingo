package outputparser

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStructured(t *testing.T) {
	t.Parallel()

	type test struct {
		name              string
		responseSchema    []ResponseSchema
		llmOutput         string
		formatInstruction string
		parsed            map[string]string
		expectError       bool
	}

	testCases := []test{
		{
			name: "Single",
			responseSchema: []ResponseSchema{
				{Name: "url", Description: "A link to the resource"},
			},
			llmOutput:         "```json\n{\n\t\"url\": \"https://google.com\" \n}\n```",
			formatInstruction: "The output should be a markdown code snippet formatted in the following schema: \n```json\n{\n\t\"url\": string // A link to the resource\n}\n```", //nolint
			parsed:            map[string]string{"url": "https://google.com"},
			expectError:       false,
		},
		{
			name: "Double",
			responseSchema: []ResponseSchema{
				{Name: "answer", Description: "The answer to the question"},
				{Name: "source", Description: "A link to the source"},
			},
			llmOutput:         " ``` foo```json \n{\n\t\"answer\": \"Paris\",\n\t\"source\": \"https://google.com\" \n}\n``` ``` bar zoo",                                                                                               //nolint
			formatInstruction: "The output should be a markdown code snippet formatted in the following schema: \n```json\n{\n\t\"answer\": string // The answer to the question\n\t\"source\": string // A link to the source\n}\n```", //nolint
			parsed:            map[string]string{"answer": "Paris", "source": "https://google.com"},
			expectError:       false,
		},
		{
			name: "MissingKey",
			responseSchema: []ResponseSchema{
				{Name: "answer", Description: "The answer to the question"},
				{Name: "source", Description: "A link to the source"},
			},
			llmOutput:         "```json \n{\n\t\"source\": \"https://google.com\" \n}\n```",
			formatInstruction: "The output should be a markdown code snippet formatted in the following schema: \n```json\n{\n\t\"answer\": string // The answer to the question\n\t\"source\": string // A link to the source\n}\n```", //nolint
			parsed:            nil,
			expectError:       true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			parser := NewStructured(tc.responseSchema)
			assert.Equal(t, tc.formatInstruction, parser.GetFormatInstructions())

			parsed, err := parser.Parse(tc.llmOutput)
			if (err != nil) != tc.expectError {
				t.Fatalf("expected error: %v, got: %v", tc.expectError, err)
			}

			assert.Equal(t, parsed, tc.parsed)
		})
	}
}
