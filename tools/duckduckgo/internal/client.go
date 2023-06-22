package internal

import (
	"errors"
	"fmt"

	"gitlab.com/psheets/ddgquery"
)

var ErrNoGoodResult = errors.New("no good search results found")

type Client struct{}

type SearchResult struct {
	Source string
	Title  string
	Info   string
	Ref    string
}

func New() *Client {
	return &Client{}
}

func (client *Client) Search(query string) (string, error) {
	search := []SearchResult{}
	numberOfResults := 5

	results, _ := ddgquery.Query(query, numberOfResults)
	if len(results) == 0 {
		return "", ErrNoGoodResult
	}

	for _, result := range results {
		search = append(search, SearchResult{
			Source: "duckduckgo.com",
			Title:  result.Title,
			Info:   result.Info,
			Ref:    result.Ref,
		})
	}

	return client.formatResults(search), nil
}

// formatResults will return a structured string with the results in the format:

// Title: Leonardo DiCaprio & Camila Morrone's Full Relationship Timeline
// Source: duckduckgo.com
// Description: Leonardo DiCaprio and Camila Morrone are pros at keeping their
// relationship under wraps from public scrutiny.
// URL: //duckduckgo.com/l/?uddg=https://www.harpersbazaar.com/celebrity/latest/

// Title: Who Is Leonardo DiCaprio Dating Now 2022? Victoria Lamas, Gigi Hadid ...
// Source: duckduckgo.com
// Description: After getting out of very highly publicized relationships, many Leo fans are asking:
// Who is Leonardo DiCaprio dating now? After launching his career as an actor in the 90s, Leo has been in many...
// URL: //duckduckgo.com/l/?uddg=https://stylecaster.com/entertainment/celebrity-news/1348452

func (client *Client) formatResults(results []SearchResult) string {
	formattedResults := ""

	for _, result := range results {
		formattedResults += fmt.Sprintf(
			"Title: %s\nSource: %s\nDescription: %s\nURL: %s\n\n",
			result.Title,
			result.Source,
			result.Info, result.Ref,
		)
	}

	return formattedResults
}
