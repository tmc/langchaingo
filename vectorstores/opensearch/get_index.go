package opensearch

import (
	"context"

	"github.com/opensearch-project/opensearch-go/opensearchapi"
)

func (s *Store) GetIndex(
	ctx context.Context,
	indexName string,
	output any,
) error {
	getIndex := opensearchapi.IndicesGetRequest{
		Index: []string{indexName},
	}
	res, err := getIndex.Do(ctx, s.client)
	if err != nil {
		return err
	}
	return handleResponse(output, res)
}
