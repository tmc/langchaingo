package queryconstructor

import (
	"fmt"

	queryconstructor_parser "github.com/tmc/langchaingo/exp/tools/queryconstructor/parser"
	"github.com/tmc/langchaingo/outputparser"
	"github.com/tmc/langchaingo/schema"
)

type Translator interface {
	Translate(structuredQuery queryconstructor_parser.StructuredFilter) (any, error)
}

type QueryConstructorParser struct {
	translator Translator
}

func NewQueryConstructorParser(translator Translator) QueryConstructorParser {

	return QueryConstructorParser{
		translator: translator,
	}

}

var _ schema.OutputParser[any] = QueryConstructorParser{}

// Parse parses the output of an LLM into a map of strings.
func (qcp QueryConstructorParser) Parse(result string) (any, error) {
	var json map[string]interface{}
	var jsonAny any
	var err error
	var filter string
	var ok bool
	var structuredFilter *queryconstructor_parser.StructuredFilter

	if jsonAny, err = outputparser.NewJSONMarkdown().Parse(result); err != nil {
		return nil, fmt.Errorf("query constructor couldn't get json %w", err)
	}

	if json, ok = jsonAny.(map[string]interface{}); !ok {
		return nil, fmt.Errorf("wrong type retuned by json markdown parser")
	}

	if filter, ok = json["filter"].(string); !ok {
		return nil, fmt.Errorf("query constructor couldn't get json")
	}

	fmt.Printf("filter: %v\n", filter)

	if structuredFilter, err = queryconstructor_parser.Parse(filter); err != nil {
		return nil, fmt.Errorf("query constructor couldn't parse query %w", err)
	}

	if structuredFilter != nil {
		fmt.Printf("parsedQuery: %v\n", structuredFilter)
	}

	return qcp.translator.Translate(*structuredFilter)
}

// ParseWithPrompt does the same as Parse.
func (qcp QueryConstructorParser) ParseWithPrompt(text string, _ schema.PromptValue) (any, error) {
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
