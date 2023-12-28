package opensearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/opensearch-project/opensearch-go/opensearchapi"
)

type IndexOption func(indexMap *map[string]interface{})

const (
	engine                       = "nmslib"
	vectorField                  = "contentVector"
	spaceType                    = "l2"
	vectorDimension              = 1536
	hnswParametersM              = 16
	hnswParametersEfConstruction = 512
	hnswParametersEfSearch       = 512
)

func (s *Store) CreateIndex(ctx context.Context, indexName string, opts ...IndexOption) (*opensearchapi.Response, error) {
	indexSchema := map[string]interface{}{
		"settings": map[string]interface{}{
			"index": map[string]interface{}{
				"knn":                      true,
				"knn.algo_param.ef_search": hnswParametersEfSearch,
			},
		},
		"mappings": map[string]interface{}{
			"properties": map[string]interface{}{
				vectorField: map[string]interface{}{
					"type":      "knn_vector",
					"dimension": vectorDimension,
					"method": map[string]interface{}{
						"name":       "hnsw",
						"space_type": spaceType,
						"engine":     engine,
						"parameters": map[string]interface{}{
							"ef_construction": hnswParametersEfConstruction,
							"m":               hnswParametersM,
						},
					},
				},
			},
		},
	}

	for _, indexOption := range opts {
		indexOption(&indexSchema)
	}

	buf := new(bytes.Buffer)

	if err := json.NewEncoder(buf).Encode(indexSchema); err != nil {
		return nil, fmt.Errorf("error encoding index schema to json buffer %w", err)
	}

	indice := opensearchapi.IndicesCreateRequest{
		Index: indexName,
		Body:  buf,
	}

	return indice.Do(ctx, s.client)
}
