package selfquery

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/outputparser"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/tools/queryconstructor"
	queryconstructor_parser "github.com/tmc/langchaingo/tools/queryconstructor/parser"
)

var ErrEmptyFilter = fmt.Errorf("empty filter")

// main function to retrieve documents with a query prompt.
func (sqr Retriever) GetRelevantDocuments(ctx context.Context, query string) ([]schema.Document, error) {
	prompt, err := queryconstructor.GetQueryConstructorPrompt(queryconstructor.GetQueryConstructorPromptArgs{
		DocumentContents: sqr.DocumentContents,
		AttributeInfo:    sqr.MetadataFieldInfo,
		EnableLimit:      &sqr.EnableLimit,
	})
	if err != nil {
		return nil, fmt.Errorf("error load query constructor %w", err)
	}

	promptChain := *chains.NewLLMChain(
		sqr.LLM,
		prompt,
		chains.WithTemperature(0),
	)

	promptChain.OutputParser = outputparser.NewJSONMarkdown()

	result, err := promptChain.Call(ctx, map[string]any{
		"query": query,
	})
	if err != nil {
		return nil, err
	}

	var resultBytes []byte
	var output map[string]interface{}
	var ok bool

	if resultBytes, ok = result["text"].([]byte); !ok {
		return nil, fmt.Errorf("wrong type retuned by json markdown parser")
	}

	if err = json.Unmarshal(resultBytes, &output); err != nil {
		return nil, fmt.Errorf("wrong json retuned by json markdown parser")
	}

	if sqr.CaptureOutput != nil {
		(*sqr.CaptureOutput)(output)
	}

	var queryRefinedPrompt string

	filters, err := sqr.parseFilter(output["filter"])
	if err != nil && !errors.Is(ErrEmptyFilter, err) {
		return nil, err
	}

	if refinedPrompt, ok := output["query"].(string); ok {
		queryRefinedPrompt = refinedPrompt
	}

	limit, _ := output["limit"].(int)

	if limit == 0 {
		limit = sqr.DefaultLimit
	}

	if sqr.Debug {
		log.Printf("query refined prompt: %s", queryRefinedPrompt)
		log.Printf("filters: %s", filters)
		log.Printf("limit: %d", limit)
	}

	return sqr.Store.Search(ctx, queryRefinedPrompt, filters, limit)
}

func (sqr Retriever) parseFilter(input interface{}) (any, error) {
	var err error

	if filter, ok := input.(string); ok && filter != "NO_FILTER" {
		if sqr.Debug {
			log.Printf("filter: %v\n", filter)
		}
		var structuredFilter *queryconstructor_parser.StructuredFilter
		if structuredFilter, err = queryconstructor_parser.Parse(filter); err != nil {
			return nil, fmt.Errorf("query constructor couldn't parse query %w", err)
		}
		return sqr.Store.Translate(*structuredFilter)
	}

	return nil, ErrEmptyFilter
}
