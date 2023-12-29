package azureaisearch

import (
	"context"
	"fmt"
	"net/http"
)

// CreateIndexAPIRequest send a request to azure AI search Rest API for deleting an index.
func (s *Store) DeleteIndex(ctx context.Context, indexName string) error {
	URL := fmt.Sprintf("%s/indexes/%s?api-version=2023-11-01", s.azureAISearchEndpoint, indexName)
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, URL, nil)
	if err != nil {
		return fmt.Errorf("err setting request for index creating: %w", err)
	}

	req.Header.Add("Content-Type", "application/json")
	if s.azureAISearchAPIKey != "" {
		req.Header.Add("api-key", s.azureAISearchAPIKey)
	}

	if err := s.httpDefaultSend(req, "index creating for azure ai search", nil); err != nil {
		return fmt.Errorf("err request: %w", err)
	}

	return nil
}
