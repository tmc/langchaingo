package queryconstructor

import queryconstructor_parser "github.com/tmc/langchaingo/exp/tools/queryconstructor/parser"

type StructuredQuery struct {
	Query   string
	Filters queryconstructor_parser.StructuredFilter
	Limit   *int
}
