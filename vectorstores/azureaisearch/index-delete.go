package azureaisearch

import (
	"context"
	"fmt"
	"net/http"
)

func (s *Store) DeleteIndex(ctx context.Context, indexName string) error {
	URL := fmt.Sprintf("%s/indexes/%s?api-version=2023-11-01", s.cognitiveSearchEndpoint, indexName)
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, URL, nil)
	if err != nil {
		return fmt.Errorf("err setting request for index creating: %w", err)
	}

	req.Header.Add("Content-Type", "application/json")
	if s.cognitiveSearchAPIKey != "" {
		req.Header.Add("api-key", s.cognitiveSearchAPIKey)
	}

	if err := s.HTTPDefaultSend(req, "index creating for cognitive search", nil); err != nil {
		return fmt.Errorf("err request: %w", err)
	}

	return nil
}
