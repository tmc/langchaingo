package outputparser

import "github.com/tmc/langchaingo/schema"

// Simple is an output parser that does nothing.
type Simple struct{}

func NewSimple() Simple { return Simple{} }

var _ schema.OutputParser[any] = Simple{}

func (p Simple) GetFormatInstructions() string  { return "" }
func (p Simple) Parse(text string) (any, error) { return text, nil }
func (p Simple) ParseWithPrompt(text string, _ schema.PromptValue) (any, error) {
	return text, nil
}
func (p Simple) Type() string { return "simple_parser" }
