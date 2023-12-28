package azureaisearch

import (
	"context"
	"fmt"

	"net/http"
)

func (s *Store) RetrieveIndex(ctx context.Context, indexName string, output *map[string]interface{}) error {
	URL := fmt.Sprintf("%s/indexes/%s?api-version=2023-11-01", s.cognitiveSearchEndpoint, indexName)
	req, err := http.NewRequest(http.MethodGet, URL, nil)

	if err != nil {
		fmt.Printf("err setting request for index retrieving: %v\n", err)
		return err
	}

	req.Header.Add("Content-Type", "application/json")
	if s.cognitiveSearchAPIKey != "" {
		req.Header.Add("api-key", s.cognitiveSearchAPIKey)
	}

	return s.HTTPDefaultSend(req, "search documents on cognitive search", output)
}
