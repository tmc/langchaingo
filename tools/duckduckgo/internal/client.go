package internal

import (
	"context"
	"errors"
	"fmt"

	"gitlab.com/psheets/ddgquery"
)

var (
	ErrNoGoodResult = errors.New("no good search results found")
)

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

func (client *Client) Search(ctx context.Context, query string) (string, error) {
	var search []SearchResult

	results, _ := ddgquery.Query(query, 5)
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

// formatResults wil return a structured string with the results in the format:

// Title: Leonardo DiCaprio & Camila Morrone's Full Relationship Timeline
// Source: duckduckgo.com
// Description: Leonardo DiCaprio and Camila Morrone are pros at keeping their relationship under wraps from public scrutiny.
// URL: //duckduckgo.com/l/?uddg=https://www.harpersbazaar.com/celebrity/latest/a30841751/leonardo-dicaprio-camila-morrone-relationship-timeline/&rut=89ab942c541611a9bce3d7d0f1b017f5f8a11599476032efa74bbf3131de9804

// Title: Who Is Leonardo DiCaprio Dating Now 2022? Victoria Lamas, Gigi Hadid ...
// Source: duckduckgo.com
// Description: After getting out of very highly publicized relationships, many Leo fans are asking: Who is Leonardo DiCaprio dating now? After launching his career as an actor in the 90s, Leo has been in many...
// URL: //duckduckgo.com/l/?uddg=https://stylecaster.com/entertainment/celebrity-news/1348452/leonardo-dicaprio-dating/&rut=8664a8f16ce77d285924fe949a698484ec48cb1f9cf3aa16fbea349e383891cc

func (client *Client) formatResults(results []SearchResult) string {
	formattedResults := ""

	for _, result := range results {
		formattedResults += fmt.Sprintf("Title: %s\nSource: %s\nDescription: %s\nURL: %s\n\n", result.Title, result.Source, result.Info, result.Ref)
	}

	return formattedResults
}
