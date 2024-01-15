package selfquery_opensearch

import (
	"context"

	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/vectorstores/opensearch"
)

func (sqt SelfQueryOpensearchTranslator) Search(ctx context.Context, query string, filters any) ([]schema.Document, error) {
	return sqt.vectorstore.SimilaritySearch(ctx, query, -1, opensearch.WithFilters(filters))
}
