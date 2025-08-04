package outputparser

import (
	"strings"
	
	"github.com/yincongcyincong/langchaingo/llms"
	"github.com/yincongcyincong/langchaingo/schema"
)

// Simple is an output parser that does nothing.
type Simple struct{}

func NewSimple() Simple { return Simple{} }

var _ schema.OutputParser[any] = Simple{}

func (p Simple) GetFormatInstructions() string { return "" }
func (p Simple) Parse(text string) (any, error) {
	return strings.TrimSpace(text), nil
}

func (p Simple) ParseWithPrompt(text string, _ llms.PromptValue) (any, error) {
	return strings.TrimSpace(text), nil
}
func (p Simple) Type() string { return "simple_parser" }
