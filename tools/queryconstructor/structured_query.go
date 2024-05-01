package queryconstructor

import queryconstructor_parser "github.com/tmc/langchaingo/tools/queryconstructor/parser"

// we want the LLM to redefine a prompt, a query for structured data and a limit, all of them are optional.
type StructuredQuery struct {
	Query   string
	Filters queryconstructor_parser.StructuredFilter
	Limit   *int
}
