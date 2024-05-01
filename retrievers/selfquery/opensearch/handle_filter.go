package selfqueryopensearch

import (
	"errors"

	queryconstructor_parser "github.com/tmc/langchaingo/tools/queryconstructor/parser"
)

func (sqt Translator) handleFilter(filter interface{}) (interface{}, error) {
	structuredFilter, isFunction := filter.(queryconstructor_parser.StructuredFilter)
	if !isFunction {
		return nil, errors.New("unknown argument type for operation")
	}

	if argOperator, ok := sqt.operatorMap[structuredFilter.FunctionName]; ok {
		return sqt.operation(argOperator, structuredFilter.Args)
	}

	if argComparator, ok := sqt.comparatorMap[structuredFilter.FunctionName]; ok {
		return sqt.comparison(argComparator, structuredFilter.Args)
	}
	return nil, errors.New("unknown function name")
}
