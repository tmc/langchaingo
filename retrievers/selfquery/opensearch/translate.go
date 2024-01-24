package selfqueryopensearch

import (
	queryconstructor_parser "github.com/tmc/langchaingo/tools/queryconstructor/parser"
)

// translate structuredfilter from queryconstructor to opensearch filters.
func (sqt Translator) Translate(structuredFilter queryconstructor_parser.StructuredFilter) (any, error) {
	filters, err := sqt.handleFilter(structuredFilter)
	if err != nil {
		return nil, err
	}

	return filters, nil
}
