package outputParsers

import "github.com/tmc/langchaingo/prompts"

type OutputParser interface {
	GetFormatInstructions() string
	Parse(text string) (any, error)
	ParseWithPrompt(text string, prompt prompts.PromptValue) (any, error)
}

type OutputParserException struct {
	Reason string
}

func (e OutputParserException) Error() string {
	return e.Reason
}
