package selfqueryopensearch

import (
	"context"

	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/vectorstores"
	"github.com/tmc/langchaingo/vectorstores/opensearch"
)

// trigger the search with filters, interface for vectorstore doesnt have universal way for filters.
func (sqt Translator) Search(ctx context.Context, query string, filters any, k int) ([]schema.Document, error) {
	return sqt.vectorstore.SimilaritySearch(ctx, query, k, opensearch.WithFilters(filters), vectorstores.WithNameSpace(sqt.indexName))
}
