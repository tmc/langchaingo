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

// LinksSearch defines a tool implementation for the Metaphor Find Similar Links.
type LinksSearch struct {
	client  *metaphor.Client
	options []metaphor.ClientOptions
}

var _ tools.Tool = &LinksSearch{}

// NewLinksSearch creates a new metaphor Search instance, that
// can be used to find similar links.
//
// It accepts an optional list of ClientOptions as parameters.
// It returns a pointer to a LinksSearch instance and an error.
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

// SetOptions sets the options for the LinksSearch tool.
//
// It takes in one or more ClientOptions parameters and assigns them to the tool's options field.
func (tool *LinksSearch) SetOptions(options ...metaphor.ClientOptions) {
	tool.options = options
}

// Name returns the name of the LinksSearch tool.
//
// No parameters.
// Returns a string.
func (tool *LinksSearch) Name() string {
	return "Metaphor Links Search"
}

// Description returns the description of the LinksSearch tool.
//
// This function does not take any parameters.
// It returns a string that describes the purpose of the LinksSearch tool.
func (tool *LinksSearch) Description() string {
	return `
	Metaphor Links Search finds similar links to the link provided.
	Input should be the url string for which you would like to find similar links`
}

// Call searches for similar links using the LinksSearch tool.
//
// ctx - the context in which the function is called.
// input - the string input used to find similar links, i.e. the url.
// Returns a string containing the formatted links and an error if any occurred.
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
