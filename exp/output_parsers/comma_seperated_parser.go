package output_parsers

import (
	"strings"

	"github.com/tmc/langchaingo/schema"
)

type CommaSeparatedListOutputParser struct{}

func NewCommaSeparatedList() CommaSeparatedListOutputParser {
	return CommaSeparatedListOutputParser{}
}

func (p CommaSeparatedListOutputParser) GetFormatInstructions() string {
	return "Your response should be a list of comma separated values, eg: `foo, bar, baz`"
}

func (p CommaSeparatedListOutputParser) Parse(text string) (any, error) {
	values := strings.Split(strings.TrimSpace(text), ",")
	for i := 0; i < len(values); i++ {
		values[i] = strings.TrimSpace(values[i])
	}

	return values, nil
}

func (p CommaSeparatedListOutputParser) ParseWithPrompt(text string, prompt schema.PromptValue) (any, error) {
	return p.Parse(text)
}
