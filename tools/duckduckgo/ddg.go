package duckduckgo

import (
	"context"
	"errors"

	"github.com/tmc/langchaingo/tools"
	"github.com/tmc/langchaingo/tools/duckduckgo/internal"
)

type Tool struct {
	client *internal.Client
}

var _ tools.Tool = Tool{}

func New() (*Tool, error) {
	maxResults := 5
	return &Tool{
		client: internal.New(maxResults),
	}, nil
}

func (t Tool) Name() string {
	return "DuckDuckGo Search"
}

func (t Tool) Description() string {
	return `
	"A wrapper around DuckDuckGo Search."
	"Free search alternative to google and serpapi."
	"Input should be a search query."`
}

func (t Tool) Call(ctx context.Context, input string) (string, error) {
	result, err := t.client.Search(ctx, input)
	if err != nil {
		if errors.Is(err, internal.ErrNoGoodResult) {
			return "No good DuckDuckGo Search Results was found", nil
		}
		return "", err
	}

	return result, nil
}
