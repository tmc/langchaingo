package selfquery

import (
	"context"
	"fmt"

	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/exp/tools/queryconstructor"
	queryconstructor_parser "github.com/tmc/langchaingo/exp/tools/queryconstructor/parser"
	"github.com/tmc/langchaingo/outputparser"
	"github.com/tmc/langchaingo/schema"
)

func (sqr SelfQueryRetriever) GetRelevantDocuments(ctx context.Context, query string) ([]schema.Document, error) {
	prompt, err := queryconstructor.GetQueryConstructorPrompt(queryconstructor.GetQueryConstructorPromptArgs{
		DocumentContents: sqr.DocumentContents,
		AttributeInfo:    sqr.MetadataFieldInfo,
		EnableLimit:      sqr.EnableLimit,
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

	var json map[string]interface{}
	var ok bool

	if json, ok = result["text"].(map[string]interface{}); !ok {
		return nil, fmt.Errorf("wrong type retuned by json markdown parser")
	}

	fmt.Printf("result: %v\n", result)

	var filters any
	var queryRefinedPrompt string

	if filter, ok := json["filter"].(string); ok && filter != "NO_FILTER" {
		if filters, err = sqr.parseFilter(filter); err != nil {
			return nil, err
		}
	}

	if refinedPrompt, ok := json["query"].(string); ok {
		queryRefinedPrompt = refinedPrompt
	}

	limit, _ := json["limit"].(int)

	if limit == 0 {
		limit = sqr.DefaultLimit
	}

	return sqr.Store.Search(ctx, queryRefinedPrompt, filters, limit)
}

func (sqr SelfQueryRetriever) parseFilter(filter string) (any, error) {
	var err error
	var structuredFilter *queryconstructor_parser.StructuredFilter
	if structuredFilter, err = queryconstructor_parser.Parse(filter); err != nil {
		return nil, fmt.Errorf("query constructor couldn't parse query %w", err)
	}

	if structuredFilter != nil {
		fmt.Printf("parsedQuery: %v\n", structuredFilter)
	}

	return sqr.Store.Translate(*structuredFilter)
}
