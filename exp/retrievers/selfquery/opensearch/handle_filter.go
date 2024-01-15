package selfquery_opensearch

import (
	"errors"

	queryconstructor_parser "github.com/tmc/langchaingo/exp/tools/queryconstructor/parser"
)

func (sqt SelfQueryOpensearchTranslator) HandleFilter(filter interface{}) (interface{}, error) {
	structuredFilter, isFunction := filter.(queryconstructor_parser.StructuredFilter)
	if !isFunction {
		return nil, errors.New("unknown argument type for operation")
	}

	if argOperator, ok := OperatorMap[structuredFilter.FunctionName]; ok {
		return sqt.Operation(argOperator, structuredFilter.Args)
	}

	if argComparator, ok := ComparatorMap[structuredFilter.FunctionName]; ok {
		return sqt.Comparison(argComparator, structuredFilter.Args)
	}
	return nil, errors.New("unknown function name")
}
