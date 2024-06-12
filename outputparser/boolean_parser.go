package outputparser

import (
	"fmt"
	"strings"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/schema"
	"golang.org/x/exp/slices"
)

// BooleanParser is an output parser used to parse the output of an LLM as a boolean.
type BooleanParser struct {
	TrueStr  string
	FalseStr string
}

// NewBooleanParser returns a new BooleanParser.
func NewBooleanParser() BooleanParser {
	return BooleanParser{
		TrueStr:  "YES",
		FalseStr: "NO",
	}
}

// Statically assert that BooleanParser implements the OutputParser interface.
var _ schema.OutputParser[any] = BooleanParser{}

// GetFormatInstructions returns instructions on the expected output format.
func (p BooleanParser) GetFormatInstructions() string {
	return "Your output should be a boolean. e.g.:\n `true` or `false`"
}

func (p BooleanParser) parse(text string) (bool, error) {
	text = normalize(text)
	booleanStrings := []string{p.TrueStr, p.FalseStr}

	if !slices.Contains(booleanStrings, text) {
		return false, ParseError{
			Text:   text,
			Reason: fmt.Sprintf("Expected output to be either '%s' or '%s', received %s", p.TrueStr, p.FalseStr, text),
		}
	}

	return text == p.TrueStr, nil
}

func normalize(text string) string {
	text = strings.TrimSpace(text)
	text = strings.ToUpper(text)

	return text
}

// Parse parses the output of an LLM into a map of strings.
func (p BooleanParser) Parse(text string) (any, error) {
	return p.parse(text)
}

// ParseWithPrompt does the same as Parse.
func (p BooleanParser) ParseWithPrompt(text string, _ llms.PromptValue) (any, error) {
	return p.parse(text)
}

// Type returns the type of the parser.
func (p BooleanParser) Type() string {
	return "boolean_parser"
}
