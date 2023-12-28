package opensearch

import (
	"context"

	"github.com/opensearch-project/opensearch-go/opensearchapi"
)

func (s *Store) DeleteIndex(ctx context.Context, indexName string, opts ...IndexOption) (*opensearchapi.Response, error) {
	deleteIndex := opensearchapi.IndicesDeleteRequest{
		Index: []string{indexName},
	}

	return deleteIndex.Do(ctx, s.client)
}
