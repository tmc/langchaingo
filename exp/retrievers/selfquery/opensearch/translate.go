package selfquery_opensearch

import (
	queryconstructor_parser "github.com/tmc/langchaingo/exp/tools/queryconstructor/parser"
)

func (sqt SelfQueryOpensearchTranslator) Translate(structuredFilter queryconstructor_parser.StructuredFilter) (any, error) {

	filters, err := sqt.HandleFilter(structuredFilter)
	if err != nil {
		return nil, err
	}

	return filters, nil

}
