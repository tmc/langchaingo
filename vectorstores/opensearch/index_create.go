package opensearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/opensearch-project/opensearch-go/opensearchapi"
)

// IndexOption for passing the schema of the index as option argument for custom modification.
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

// CreateIndex for creating an index before to add a document to it.
func (s *Store) CreateIndex(
	ctx context.Context,
	indexName string,
	output any,
	opts ...IndexOption,
) error {
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
				"metadata": map[string]interface{}{},
			},
		},
	}

	for _, indexOption := range opts {
		indexOption(&indexSchema)
	}

	buf := new(bytes.Buffer)

	if err := json.NewEncoder(buf).Encode(indexSchema); err != nil {
		return fmt.Errorf("error encoding index schema to json buffer %w", err)
	}

	indice := opensearchapi.IndicesCreateRequest{
		Index: indexName,
		Body:  buf,
	}
	res, err := indice.Do(ctx, s.client)
	if err != nil {
		return err
	}
	return handleResponse(output, res)
}

func WithMetadata(metadata any) IndexOption {
	return func(indexMap *map[string]interface{}) {
		if mappings, ok := (*indexMap)["mappings"].(map[string]interface{}); ok {
			if properties, ok := mappings["properties"].(map[string]interface{}); ok {
				if metadataMap, ok := properties["metadata"].(map[string]interface{}); ok {
					metadataMap["properties"] = metadata
				}
			}
		}
	}
}
