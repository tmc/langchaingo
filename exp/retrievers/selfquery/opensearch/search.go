package selfquery_opensearch

import (
	"context"
	"fmt"

	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/vectorstores/opensearch"
)

func (sqt SelfQueryOpensearchTranslator) Search(ctx context.Context, query string, filters any) ([]schema.Document, error) {
	fmt.Printf("sqt: %+v\n", sqt)
	return sqt.vectorstore.SimilaritySearch(ctx, query, 4, opensearch.WithFilters(filters))
}
