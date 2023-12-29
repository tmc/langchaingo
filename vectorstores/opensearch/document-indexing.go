package opensearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/opensearch-project/opensearch-go/opensearchapi"
)

type Document struct {
	FieldsContent       string    `json:"content"`
	FieldsContentVector []float32 `json:"contentVector"`
	FieldsMetadata      string    `json:"metadata"`
}

func (s *Store) DocumentIndexing(
	ctx context.Context,
	indexName string,
	text string,
	vector []float32,
	metadata map[string]any,
) (*opensearchapi.Response, error) {
	metadataString, err := json.Marshal(metadata)
	if err != nil {
		return nil, err
	}

	document := Document{
		FieldsContent:       text,
		FieldsContentVector: vector,
		FieldsMetadata:      string(metadataString),
	}

	buf := new(bytes.Buffer)

	if err := json.NewEncoder(buf).Encode(document); err != nil {
		return nil, fmt.Errorf("error encoding index schema to json buffer %w", err)
	}

	indice := opensearchapi.IndexRequest{
		Index:      indexName,
		DocumentID: uuid.NewString(),
		Body:       buf,
	}

	return indice.Do(ctx, s.client)
}
