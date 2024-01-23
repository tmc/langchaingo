package selfquery_opensearch

import (
	"context"

	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/vectorstores"
	"github.com/tmc/langchaingo/vectorstores/opensearch"
)

func (sqt SelfQueryOpensearchTranslator) Search(ctx context.Context, query string, filters any, k int) ([]schema.Document, error) {
	return sqt.vectorstore.SimilaritySearch(ctx, query, k, opensearch.WithFilters(filters), vectorstores.WithNameSpace(sqt.indexName))
}
