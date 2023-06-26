package internal

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

var (
	ErrNoGoodResult = errors.New("no good search results found")
	ErrAPIResponse  = errors.New("duckduckgo api responded with error")
)

type Client struct {
	maxResults int
}

type Result struct {
	Title string
	Info  string
	Ref   string
}

func New(maxResults int) *Client {
	return &Client{
		maxResults: maxResults,
	}
}

func (client *Client) Search(ctx context.Context, query string) (string, error) {
	results := []Result{}
	queryURL := fmt.Sprintf("https://duckduckgo.com/html/?q=%s", url.QueryEscape(query))

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, queryURL, nil)
	if err != nil {
		return "", fmt.Errorf("creating duckduckgo request: %w", err)
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return "", fmt.Errorf("get %s error: %w", queryURL, err)
	}

	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return "", ErrAPIResponse
	}

	doc, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		return "", fmt.Errorf("new document error: %w", err)
	}

	sel := doc.Find(".web-result")

	for i := range sel.Nodes {
		// Break loop once required amount of results are add
		if client.maxResults == len(results) {
			break
		}
		node := sel.Eq(i)
		titleNode := node.Find(".result__a")

		info := node.Find(".result__snippet").Text()
		title := titleNode.Text()

		ref, err := url.QueryUnescape(
			strings.TrimPrefix(
				titleNode.Nodes[0].Attr[2].Val,
				"/l/?kh=-1&uddg=",
			),
		)
		if err != nil {
			log.Printf("Error: %s", err)
			return "", err
		}

		results = append(results, Result{title, info, ref})
	}

	return client.formatResults(results), nil
}

func (client *Client) SetMaxResults(n int) {
	client.maxResults = n
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

func (client *Client) formatResults(results []Result) string {
	formattedResults := ""

	for _, result := range results {
		formattedResults += fmt.Sprintf("Title: %s\nDescription: %s\nURL: %s\n\n", result.Title, result.Info, result.Ref)
	}

	return formattedResults
}
