package opensearch

import (
	"context"

	"github.com/opensearch-project/opensearch-go/opensearchapi"
)

// DeleteIndex for deleting an index before to add a document to it
func (s *Store) DeleteIndex(
	ctx context.Context,
	indexName string,
) (*opensearchapi.Response, error) {
	deleteIndex := opensearchapi.IndicesDeleteRequest{
		Index: []string{indexName},
	}

	return deleteIndex.Do(ctx, s.client)
}
