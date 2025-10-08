package outputparser

import (
	"fmt"
	"regexp"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/schema"
)

// RegexDict is an output parser used to parse the output of an LLM as a map.
type RegexDict struct {
	// OutputKeyToFormat is a map which has a key that represents an identifier for the regex,
	// and a value to search for as a key in the parsed text.
	OutputKeyToFormat map[string]string

	// NoUpdateValue is a string that prevents assignment to parsed outputs map when matched.
	NoUpdateValue string
}

const (
	regexDictPattern = `(?:%s):\s?(?P<value>(?:[^.'\n']*)\.?)`
)

// NewRegexDict returns a new RegexDict.
func NewRegexDict(outputKeyToFormat map[string]string, noUpdateValue string) RegexDict {
	return RegexDict{
		OutputKeyToFormat: outputKeyToFormat,
		NoUpdateValue:     noUpdateValue,
	}
}

// Statically assert that RegexDict implements the OutputParser interface.
var _ schema.OutputParser[any] = RegexDict{}

// GetFormatInstructions returns instructions on the expected output format.
func (p RegexDict) GetFormatInstructions() string {
	instructions := "Your output should be a map of strings. e.g.:\n"
	instructions += "map[string]string{\"key1\": \"value1\", \"key2\": \"value2\"}\n"

	return instructions
}

func (p RegexDict) parse(text string) (map[string]string, error) {
	results := make(map[string]string, len(p.OutputKeyToFormat))
	// We only expect to get a single matched value pair for each output key.
	expectedMatches := 2

	for key, format := range p.OutputKeyToFormat {
		expression := regexp.MustCompile(fmt.Sprintf(regexDictPattern, format))
		matches := expression.FindStringSubmatch(text)

		if len(matches) < expectedMatches {
			return nil, ParseError{
				Text:   text,
				Reason: fmt.Sprintf("No match found for expression %s", expression),
			}
		}

		if len(matches) > expectedMatches {
			return nil, ParseError{
				Text:   text,
				Reason: fmt.Sprintf("Multiple matches found for expression %s", expression),
			}
		}

		match := matches[1]

		if match == p.NoUpdateValue {
			continue
		}

		results[key] = match
	}

	return results, nil
}

// Parse parses the output of an LLM into a map of strings.
func (p RegexDict) Parse(text string) (any, error) {
	return p.parse(text)
}

// ParseWithPrompt does the same as Parse.
func (p RegexDict) ParseWithPrompt(text string, _ llms.PromptValue) (any, error) {
	return p.parse(text)
}

// Type returns the type of the parser.
func (p RegexDict) Type() string {
	return "regex_dict_parser"
}
