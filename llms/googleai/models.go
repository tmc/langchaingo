package googleai

import (
	"context"
	"fmt"
	"strings"
)

// ListModels returns the model ids the endpoint serves, with the resource
// prefix stripped so the names match the ones callers pass as a model.
func (g *GoogleAI) ListModels(ctx context.Context) ([]string, error) {
	var ids []string
	for model, err := range g.client.Models.All(ctx) {
		if err != nil {
			return nil, fmt.Errorf("googleai: list models: %w", err)
		}
		if model == nil {
			continue
		}
		if id := strings.TrimPrefix(model.Name, "models/"); id != "" {
			ids = append(ids, id)
		}
	}
	return ids, nil
}
