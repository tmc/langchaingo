package outputparser

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"

	"github.com/tmc/langchaingo/schema"
)

// Simple is an output parser that does nothing.
type JSONMarkdown struct{}

func NewJSONMarkdown() JSONMarkdown { return JSONMarkdown{} }

var _ schema.OutputParser[any] = JSONMarkdown{}

func (p JSONMarkdown) GetFormatInstructions() string { return "" }
func (p JSONMarkdown) Parse(text string) (any, error) {
	output := map[string]interface{}{}
	r := regexp.MustCompile("(?s)```json(.+)```")
	fmt.Printf("text: %v\n", text)
	result := r.FindSubmatch([]byte(text))
	if len(result) < 2 {
		return nil, errors.New("couldn't find JSON markdown")
	}

	if err := json.Unmarshal(result[1], &output); err != nil {
		fmt.Printf("result[1]: %v\n", result[1])
		return nil, fmt.Errorf("unmarshalling JSON in JSON Markdown output parser %w", err)
	}

	return output, nil
}

func (p JSONMarkdown) ParseWithPrompt(text string, _ schema.PromptValue) (any, error) {
	return p.Parse(text)
}
func (p JSONMarkdown) Type() string { return "json_markdown" }
