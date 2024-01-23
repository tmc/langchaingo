package opensearch

import (
	"context"

	"github.com/opensearch-project/opensearch-go/opensearchapi"
)

// DeleteIndex for deleting an index before to add a document to it.
func (s *Store) DeleteIndex(
	ctx context.Context,
	indexName string,
) ([]byte, error) {
	deleteIndex := opensearchapi.IndicesDeleteRequest{
		Index: []string{indexName},
	}

	return handleResponse(deleteIndex.Do(ctx, s.client))
}
