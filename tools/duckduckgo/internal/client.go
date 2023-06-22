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

func (client *Client) formatResults(results []SearchResult) string {
	formattedResults := ""

	for _, result := range results {
		formattedResults += fmt.Sprintf("Title: %s\nSource: %s\nDescription: %s\nURL: %s\n\n", result.Title, result.Source, result.Info, result.Ref)
	}

	return formattedResults
}
