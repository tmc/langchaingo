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

const _defaultUserAgent = "github.com/tmc/langchaingo/tools/duckduckgo"

var (
	ErrNoGoodResult = errors.New("no good search results found")
	ErrAPIResponse  = errors.New("duckduckgo api responded with error")
)

type Client struct {
	maxResults int
	userAgent  string
}

type Result struct {
	Title string
	Info  string
	Ref   string
}

func New(maxResults int, userAgent string) *Client {
	if userAgent == "" {
		userAgent = _defaultUserAgent
	}
	if maxResults == 0 {
		maxResults = 1
	}

	return &Client{
		maxResults: maxResults,
		userAgent:  userAgent,
	}
}

func (client *Client) Search(ctx context.Context, query string) (string, error) {
	results := []Result{}
	queryURL := fmt.Sprintf("https://html.duckduckgo.com/html/?q=%s", url.QueryEscape(query))

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, queryURL, nil)
	if err != nil {
		return "", fmt.Errorf("creating duckduckgo request: %w", err)
	}

	request.Header.Add("User-Agent", client.userAgent)

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
		ref := ""

		if len(titleNode.Nodes) > 0 && len(titleNode.Nodes[0].Attr) > 2 {
			ref, err = url.QueryUnescape(
				strings.TrimPrefix(
					titleNode.Nodes[0].Attr[2].Val,
					"/l/?kh=-1&uddg=",
				),
			)
			if err != nil {
				log.Printf("Error: %s", err)
				return "", err
			}
		}

		results = append(results, Result{title, info, ref})
	}

	return client.formatResults(results), nil
}

func (client *Client) SetMaxResults(n int) {
	client.maxResults = n
}

// formatResults will return a structured string with the results.
func (client *Client) formatResults(results []Result) string {
	formattedResults := ""

	for _, result := range results {
		formattedResults += fmt.Sprintf("Title: %s\nDescription: %s\nURL: %s\n\n", result.Title, result.Info, result.Ref)
	}

	return formattedResults
}
