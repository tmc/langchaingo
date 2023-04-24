package output_parsers

import "github.com/tmc/langchaingo/schema"

type OutputParser interface {
	GetFormatInstructions() string
	Parse(text string) (any, error)
	ParseWithPrompt(text string, prompt schema.PromptValue) (any, error)
}

type OutputParserException struct {
	Reason string
}

func (e OutputParserException) Error() string {
	return e.Reason
}
