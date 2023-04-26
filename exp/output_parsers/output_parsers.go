package output_parsers

import "github.com/tmc/langchaingo/schema"

type OutputParser interface {
	Parse(text string) (any, error)
	ParseWithPrompt(text string, prompt schema.PromptValue) (any, error)
	GetFormatInstructions() string
}

type OutputParserException struct {
	Reason string
}

func (e OutputParserException) Error() string {
	return e.Reason
}
