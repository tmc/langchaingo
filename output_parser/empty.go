package output_parser

import "github.com/tmc/langchaingo/schema"

// Empty is an output parser that does nothing.
type Empty struct{}

func NewEmptyOutputParser() Empty { return Empty{} }

var _ schema.OutputParser[string] = Empty{}

func (p Empty) GetFormatInstructions() string     { return "" }
func (p Empty) Parse(text string) (string, error) { return text, nil }
func (p Empty) ParseWithPrompt(text string, prompt schema.PromptValue) (string, error) {
	return text, nil
}
func (p Empty) Type() string { return "empty_parser" }
