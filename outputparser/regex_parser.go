package outputparser

import (
	"fmt"
	"regexp"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/schema"
)

// RegexParser is an output parser used to parse the output of an LLM as a map.
type RegexParser struct {
	Expression *regexp.Regexp
	OutputKeys []string
}

// NewRegexParser returns a new RegexParser.
func NewRegexParser(expressionStr string) RegexParser {
	expression := regexp.MustCompile(expressionStr)
	outputKeys := expression.SubexpNames()[1:]

	return RegexParser{
		Expression: expression,
		OutputKeys: outputKeys,
	}
}

// Statically assert that RegexParser implements the OutputParser interface.
var _ schema.OutputParser[any] = RegexParser{}

// GetFormatInstructions returns instructions on the expected output format.
func (p RegexParser) GetFormatInstructions() string {
	instructions := "Your output should be a map of strings. e.g.:\n"
	instructions += "map[string]string{\"key1\": \"value1\", \"key2\": \"value2\"}"

	return instructions
}

func (p RegexParser) parse(text string) (map[string]string, error) {
	match := p.Expression.FindStringSubmatch(text)

	if len(match) == 0 {
		return nil, ParseError{
			Text:   text,
			Reason: fmt.Sprintf("No match found for expression %s", p.Expression),
		}
	}

	// remove the first match (entire string) for parity with the output keys.
	match = match[1:]

	matches := make(map[string]string, len(match))

	for i, name := range p.OutputKeys {
		matches[name] = match[i]
	}

	return matches, nil
}

// Parse parses the output of an LLM into a map of strings.
func (p RegexParser) Parse(text string) (any, error) {
	return p.parse(text)
}

// ParseWithPrompt does the same as Parse.
func (p RegexParser) ParseWithPrompt(text string, _ llms.PromptValue) (any, error) {
	return p.parse(text)
}

// Type returns the type of the parser.
func (p RegexParser) Type() string {
	return "regex_parser"
}
