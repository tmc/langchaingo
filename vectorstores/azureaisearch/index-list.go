package azureaisearch

import (
	"context"
	"fmt"
	"net/http"
)

func (s *Store) ListIndexes(ctx context.Context, output *map[string]interface{}) error {
	URL := fmt.Sprintf("%s/indexes?api-version=2023-11-01", s.cognitiveSearchEndpoint)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, URL, nil)
	if err != nil {
		return fmt.Errorf("err setting request for index retrieving: %w", err)
	}

	req.Header.Add("Content-Type", "application/json")
	if s.cognitiveSearchAPIKey != "" {
		req.Header.Add("api-key", s.cognitiveSearchAPIKey)
	}

	return s.HTTPDefaultSend(req, "search documents on cognitive search", output)
}
