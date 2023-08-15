// trunk-ignore(golangci-lint/dupl)
package metaphor

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/metaphorsystems/metaphor-go"
	"github.com/tmc/langchaingo/tools"
)

type LinksSearch struct {
	client  *metaphor.Client
	options []metaphor.ClientOptions
}

var _ tools.Tool = &LinksSearch{}

func NewLinksSearch(options ...metaphor.ClientOptions) (*LinksSearch, error) {
	apiKey := os.Getenv("METAPHOR_API_KEY")

	client, err := metaphor.NewClient(apiKey, options...)
	if err != nil {
		return nil, err
	}
	metaphor := &LinksSearch{
		client:  client,
		options: options,
	}

	return metaphor, nil
}

func (tool *LinksSearch) SetOptions(options ...metaphor.ClientOptions) {
	tool.options = options
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
	links, err := tool.client.FindSimilar(ctx, input, tool.options...)
	if err != nil {
		if errors.Is(err, metaphor.ErrNoLinksFound) {
			return "Metaphor Links Search didn't return any results", nil
		}
		return "", err
	}

	return tool.formatLinks(links), nil
}

func (tool *LinksSearch) formatLinks(response *metaphor.SearchResponse) string {
	formattedResults := ""

	for _, result := range response.Results {
		formattedResults += fmt.Sprintf("Title: %s\nURL: %s\nID: %s\n\n", result.Title, result.URL, result.ID)
	}

	return formattedResults
}
