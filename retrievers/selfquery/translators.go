package selfquery

import (
	"context"

	"github.com/tmc/langchaingo/schema"
	queryconstructor_parser "github.com/tmc/langchaingo/tools/queryconstructor/parser"
)

// SelfQuery needs to translate and search with filters.
type StoreWithQueryTranslator interface {
	Translate(structuredQuery queryconstructor_parser.StructuredFilter) (any, error)
	Search(ctx context.Context, query string, filters any, k int) ([]schema.Document, error)
}
