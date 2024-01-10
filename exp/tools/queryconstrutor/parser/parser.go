package queryconstructor_parser

import (
	"fmt"

	"github.com/tmc/langchaingo/schema"
)

type QueryConstructorParser struct {
}

var _ schema.OutputParser[*Function] = QueryConstructorParser{}

// Parse parses the output of an LLM into a map of strings.
func (qcp QueryConstructorParser) Parse(text string) (*Function, error) {

	lexer := NewLexer(text)
	yyParse(&lexer)

	fmt.Printf("lexer.function: %v\n", lexer.function)

	if lexer.err != nil {
		return nil, lexer.err
	}

	return &lexer.function, nil
}

// ParseWithPrompt does the same as Parse.
func (qcp QueryConstructorParser) ParseWithPrompt(text string, _ schema.PromptValue) (*Function, error) {
	return qcp.Parse(text)
}

// GetFormatInstructions returns instructions on the expected output format.
func (qcp QueryConstructorParser) GetFormatInstructions() string {
	return ""
}

// Type returns the type of the parser.
func (qcp QueryConstructorParser) Type() string {
	return "queryconstructor_parser"
}
