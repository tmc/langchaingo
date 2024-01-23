package selfquery

import (
	"context"

	queryconstructor_parser "github.com/tmc/langchaingo/exp/tools/queryconstructor/parser"
	"github.com/tmc/langchaingo/schema"
)

// SelfQuery needs to translate and search with filters
type StoreWithQueryTranslator interface {
	Translate(structuredQuery queryconstructor_parser.StructuredFilter) (any, error)
	Search(ctx context.Context, query string, filters any, k int) ([]schema.Document, error)
}
