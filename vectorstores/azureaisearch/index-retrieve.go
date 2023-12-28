package azureaisearch

import (
	"context"
	"fmt"
	"net/http"
)

func (s *Store) RetrieveIndex(ctx context.Context, indexName string, output *map[string]interface{}) error {
	URL := fmt.Sprintf("%s/indexes/%s?api-version=2023-11-01", s.azureAISearchEndpoint, indexName)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, URL, nil)
	if err != nil {
		return fmt.Errorf("err setting request for index retrieving: %w", err)
	}

	req.Header.Add("Content-Type", "application/json")
	if s.azureAISearchAPIKey != "" {
		req.Header.Add("api-key", s.azureAISearchAPIKey)
	}

	return s.HTTPDefaultSend(req, "search documents on azure ai search", output)
}
