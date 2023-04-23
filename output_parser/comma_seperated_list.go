package output_parser

import (
	"strings"

	"github.com/tmc/langchaingo/schema"
)

// CommaSeparatedList is an output parser used to parse the output of an llm as a slice of strings.
type CommaSeparatedList struct{}

// NewCommaSeparatedList creates a new CommaSeparatedList.
func NewCommaSeparatedList() CommaSeparatedList {
	return CommaSeparatedList{}
}

// Statically assert that CommaSeparatedList implement the OutputParser interface.
var _ schema.OutputParser[[]string] = CommaSeparatedList{}

func (p CommaSeparatedList) GetFormatInstructions() string {
	return "Your response should be a list of comma separated values, eg: `foo, bar, baz`"
}

func (p CommaSeparatedList) Parse(text string) ([]string, error) {
	values := strings.Split(strings.TrimSpace(text), ",")
	for i := 0; i < len(values); i++ {
		values[i] = strings.TrimSpace(values[i])
	}

	return values, nil
}

func (p CommaSeparatedList) ParseWithPrompt(text string, prompt schema.PromptValue) ([]string, error) {
	return p.Parse(text)
}

func (p CommaSeparatedList) Type() string {
	return "comma_separated_list_parser"
}
