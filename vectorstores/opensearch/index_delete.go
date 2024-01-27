package opensearch

import (
	"context"

	"github.com/opensearch-project/opensearch-go/opensearchapi"
)

// DeleteIndex for deleting an index before to add a document to it.
func (s *Store) DeleteIndex(
	ctx context.Context,
	indexName string,
	output any,
) error {
	deleteIndex := opensearchapi.IndicesDeleteRequest{
		Index: []string{indexName},
	}
	res, err := deleteIndex.Do(ctx, s.client)
	if err != nil {
		return err
	}
	return handleResponse(output, res)
}
