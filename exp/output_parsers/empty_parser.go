package output_parsers

import "github.com/tmc/langchaingo/schema"

// Output parser that does nothing.
type EmptyOutputParser struct{}

func (p EmptyOutputParser) GetFormatInstructions() string  { return "" }
func (p EmptyOutputParser) Parse(text string) (any, error) { return text, nil }
func (p EmptyOutputParser) ParseWithPrompt(text string, prompt schema.PromptValue) (any, error) {
	return text, nil
}

func NewEmptyOutputParser() EmptyOutputParser {
	return EmptyOutputParser{}
}
