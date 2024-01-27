package opensearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/opensearch-project/opensearch-go/opensearchapi"
)

type document struct {
	FieldsContent       string                 `json:"content"`
	FieldsContentVector []float32              `json:"contentVector"`
	FieldsMetadata      map[string]interface{} `json:"metadata"`
}

func (s *Store) documentIndexing(
	ctx context.Context,
	id string,
	indexName string,
	text string,
	vector []float32,
	metadata map[string]any,
	output any,
) error {
	document := document{
		FieldsContent:       text,
		FieldsContentVector: vector,
		FieldsMetadata:      metadata,
	}

	buf := new(bytes.Buffer)

	if err := json.NewEncoder(buf).Encode(document); err != nil {
		return fmt.Errorf("error encoding index schema to json buffer %w", err)
	}

	indice := opensearchapi.IndexRequest{
		Index:      indexName,
		DocumentID: id,
		Body:       buf,
	}
	res, err := indice.Do(ctx, s.client)
	if err != nil {
		return err
	}
	return handleResponse(output, res)
}
