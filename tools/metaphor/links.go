package metaphor

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/tmc/langchaingo/tools"
	"github.com/tmc/langchaingo/tools/metaphor/client"
)

type LinksSearch struct {
	client *client.MetaphorClient
}

var _ tools.Tool = &LinksSearch{}

func NewLinksSearch(options ...client.Options) (*LinksSearch, error) {
	apiKey := os.Getenv("METAPHOR_API_KEY")
	if apiKey == "" {
		return nil, ErrMissingToken
	}

	client, err := client.New(apiKey, options...)
	if err != nil {
		return nil, err
	}
	metaphor := &LinksSearch{
		client: client,
	}

	return metaphor, nil
}

func (tool *LinksSearch) Name() string {
	return "Metaphor Links Search"
}

func (tool *LinksSearch) Description() string {
	return `
	Metaphor Links Search finds similar links to the link provided.
	Input should be the url string for which you would like to find similar links`
}

func (tool *LinksSearch) Call(ctx context.Context, input string) (string, error) {
	links, err := tool.client.FindSimilar(ctx, input)
	if err != nil {
		if errors.Is(err, client.ErrNoLinksFound) {
			return "Metaphor Links Search didn't return any results", nil
		}
		return "", err
	}

	return tool.formatLinks(links), nil
}

func (tool *LinksSearch) formatLinks(response *client.SearchResponse) string {
	formattedResults := ""

	for _, result := range response.Results {
		formattedResults += fmt.Sprintf("Title: %s\nURL: %s\nID: %s\n\n", result.Title, result.URL, result.ID)
	}

	return formattedResults
}
