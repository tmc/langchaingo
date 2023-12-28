package azureaisearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
)

// mimicking langchain: https://github.com/langchain-ai/langchain/blob/bfc12a4a7644cfc4d832cc4023086a7a5374f46a/libs/langchain/langchain/vectorstores/azuresearch.py
type Document struct {
	SearchAction        string    `json:"@search.action"`
	FieldsID            string    `json:"id"`
	FieldsContent       string    `json:"content"`
	FieldsContentVector []float32 `json:"contentVector"`
	FieldsMetadata      string    `json:"metadata"`
}

func (s *Store) UploadDocument(ctx context.Context, indexName string, text string, vector []float32, metadata map[string]any) error {

	metadataString, err := json.Marshal(metadata)
	if err != nil {
		fmt.Printf("error marshalling metadata: %v\n", err)
		return err
	}

	document := Document{
		SearchAction:        "upload",
		FieldsID:            uuid.NewString(),
		FieldsContent:       text,
		FieldsContentVector: vector,
		FieldsMetadata:      string(metadataString),
	}

	return s.UploadDocumentAPIRequest(ctx, indexName, document)
}

// tech debt: should use SDK when available: https://azure.github.io/azure-sdk/releases/latest/go.html
func (s *Store) UploadDocumentAPIRequest(ctx context.Context, indexName string, document any) error {
	URL := fmt.Sprintf("%s/indexes/%s/docs/index?api-version=2020-06-30", s.cognitiveSearchEndpoint, indexName)

	documentMap := map[string]interface{}{}
	err := StructToMap(document, &documentMap)
	if err != nil {
		fmt.Printf("err converting document struc to map: %v\n", err)
	}

	documentMap["@search.action"] = "mergeOrUpload"

	body, err := json.Marshal(map[string]interface{}{
		"value": []map[string]interface{}{
			documentMap,
		},
	})

	if err != nil {
		fmt.Printf("err marshalling body for cognitive search: %v\n", err)
		return err
	}

	req, err := http.NewRequest(http.MethodPost, URL, bytes.NewBuffer(body))

	if err != nil {
		fmt.Printf("err setting request for cognitive search upload document: %v\n", err)
		return err
	}

	req.Header.Add("Content-Type", "application/json")
	if s.cognitiveSearchAPIKey != "" {
		req.Header.Add("api-key", s.cognitiveSearchAPIKey)
	}

	return s.HTTPDefaultSend(req, "cognitive search upload document", nil)
}
