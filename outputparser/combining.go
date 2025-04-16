package outputparser

import (
	"fmt"
	"strings"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/schema"
)

// Combining is a parser that combines multiple parsers into one.
type Combining struct {
	Parsers []schema.OutputParser[any]
}

// NewCombining creates a new combining parser.
func NewCombining(parsers []schema.OutputParser[any]) Combining {
	p := make([]schema.OutputParser[any], len(parsers))

	copy(p, parsers)

	return Combining{
		Parsers: p,
	}
}

// Statically assert that Combining implements the OutputParser interface.
var _ schema.OutputParser[any] = Combining{}

// GetFormatInstructions returns the format instructions.
func (p Combining) GetFormatInstructions() string {
	text := "Your response will be a map of strings combining the output of parsers\n"
	text += "using text delimited by two successive newline characters, to the respective parser.\n\n"
	text += "The output parser instructions are:"

	for _, parser := range p.Parsers {
		text += "\n- " + parser.GetFormatInstructions()
	}

	return text
}

func (p Combining) parse(text string) (map[string]any, error) {
	texts := strings.Split(text, "\n\n")
	output := make(map[string]any)

	if len(p.Parsers) <= 1 {
		return nil, ParseError{
			Text:   text,
			Reason: fmt.Sprintf("Combining parser requires at least 2 parsers, got %d", len(p.Parsers)),
		}
	}

	if len(texts) != len(p.Parsers) {
		return nil, ParseError{
			Text:   text,
			Reason: fmt.Sprintf("Texts count (%d) does not match parsers count (%d)", len(texts), len(p.Parsers)),
		}
	}

	for i, textChunk := range texts {
		textChunk = strings.TrimSpace(textChunk)
		parser := p.Parsers[i]

		parsed, err := parser.Parse(textChunk)
		if err != nil {
			return nil, err
		}

		parsedMap, ok := parsed.(map[string]string)
		if !ok {
			return nil, ParseError{
				Text:   textChunk,
				Reason: fmt.Sprintf("Parser %d does not return a map of strings", parsed),
			}
		}

		for k, result := range parsedMap {
			output[k] = result
		}
	}

	return output, nil
}

// Parse parses text delimited by two successive newline (`\n\n`) characters to the respective output parsers.
func (p Combining) Parse(text string) (any, error) {
	return p.parse(text)
}

// ParseWithPrompt with prompts does the same as Parse.
func (p Combining) ParseWithPrompt(text string, _ llms.PromptValue) (any, error) {
	return p.parse(text)
}

// Type returns the type of the parser.
func (p Combining) Type() string {
	return "combining_parser"
}
