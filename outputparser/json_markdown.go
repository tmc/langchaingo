package outputparser

import (
	"errors"
	"regexp"

	"github.com/tmc/langchaingo/schema"
)

// Simple is an output parser that does nothing.
type JSONMarkdown struct{}

func NewJSONMarkdown() JSONMarkdown { return JSONMarkdown{} }

var _ schema.OutputParser[any] = JSONMarkdown{}

func (p JSONMarkdown) GetFormatInstructions() string { return "" }
func (p JSONMarkdown) Parse(text string) (any, error) {

	r := regexp.MustCompile("(?s)```json(.+)```")

	result := r.FindSubmatch([]byte(text))
	if len(result) < 2 {
		return nil, errors.New("couldn't find JSON markdown")
	}

	return result[1], nil
}

func (p JSONMarkdown) ParseWithPrompt(text string, _ schema.PromptValue) (any, error) {
	return p.Parse(text)
}
func (p JSONMarkdown) Type() string { return "json_markdown" }
