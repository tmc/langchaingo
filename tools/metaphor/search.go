//nolint:dupl
package metaphor

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/metaphorsystems/metaphor-go"
	"github.com/tmc/langchaingo/tools"
)

type Search struct {
	client  *metaphor.Client
	options []metaphor.ClientOptions
}

var _ tools.Tool = &Search{}

func NewSearch(options ...metaphor.ClientOptions) (*Search, error) {
	apiKey := os.Getenv("METAPHOR_API_KEY")

	client, err := metaphor.NewClient(apiKey, options...)
	if err != nil {
		return nil, err
	}

	metaphor := &Search{
		client:  client,
		options: options,
	}

	return metaphor, nil
}

func (tool *Search) SetOptions(options ...metaphor.ClientOptions) {
	tool.options = options
}

func (tool *Search) Name() string {
	return "Metaphor Search"
}

func (tool *Search) Description() string {
	return `
	Metaphor Search uses a transformer architecture to predict links given text,
	and it gets its power from having been trained on the way that people talk
	about links on the Internet. The model does expect queries that look like
	how people describe a link on the Internet. For example:
	"'best restaurants in SF" is a bad query, whereas
	"Here is the best restaurant in SF:" is a good query.
	`
}

func (tool *Search) Call(ctx context.Context, input string) (string, error) {
	response, err := tool.client.Search(ctx, input, tool.options...)
	if err != nil {
		if errors.Is(err, metaphor.ErrNoSearchResults) {
			return "Metaphor Search didn't return any results", nil
		}
		return "", err
	}

	return tool.formatResults(response), nil
}

func (tool *Search) formatResults(response *metaphor.SearchResponse) string {
	formattedResults := ""

	for _, result := range response.Results {
		formattedResults += fmt.Sprintf("Title: %s\nURL: %s\nID: %s\n\n", result.Title, result.URL, result.ID)
	}

	return formattedResults
}
