package internal

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// Client defines an HTTP client for communicating with arXiv.
type Client struct {
	maxResults int
	userAgent  string
}

// Result defines a search query result type.
type Result struct {
	Title       string
	Authors     []string
	Summary     string
	PdfURL      string
	PublishedAt string
}

var (
	ErrNoGoodResult = errors.New("no good search results found")
	ErrAPIResponse  = errors.New("arXiv api responded with error")
)

// NewClient initializes a Client with arguments for setting a max
// results per search query and a value for the user agent header.
func NewClient(maxResults int, userAgent string) *Client {
	if maxResults == 0 {
		maxResults = 1
	}

	return &Client{
		maxResults: maxResults,
		userAgent:  userAgent,
	}
}

func (client *Client) newRequest(ctx context.Context, queryURL string) (*http.Request, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, queryURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating arXiv request: %w", err)
	}

	if client.userAgent != "" {
		request.Header.Add("User-Agent", client.userAgent)
	}

	return request, nil
}

// Search performs a search query and returns
// the result as string and an error if any.
func (client *Client) Search(ctx context.Context, query string) (string, error) {
	queryURL := fmt.Sprintf("https://export.arxiv.org/api/query?search_query=%s&start=0&max_results=%d",
		url.QueryEscape(query), client.maxResults)

	request, err := client.newRequest(ctx, queryURL)
	if err != nil {
		return "", err
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return "", fmt.Errorf("get %s error: %w", queryURL, err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return "", ErrAPIResponse
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return "", fmt.Errorf("reading response body: %w", err)
	}

	var feed struct {
		Entries []struct {
			Title     string `xml:"title"`
			Summary   string `xml:"summary"`
			Published string `xml:"published"`
			Authors   []struct {
				Name string `xml:"name"`
			} `xml:"author"`
			Link []struct {
				Href string `xml:"href,attr"`
				Type string `xml:"type,attr"`
			} `xml:"link"`
		} `xml:"entry"`
	}

	if err := xml.Unmarshal(body, &feed); err != nil {
		return "", fmt.Errorf("unmarshaling XML: %w", err)
	}

	results := []Result{}
	for _, entry := range feed.Entries {
		authors := []string{}
		for _, author := range entry.Authors {
			authors = append(authors, author.Name)
		}

		pdfURL := ""
		for _, link := range entry.Link {
			if link.Type == "application/pdf" {
				pdfURL = link.Href
				break
			}
		}

		results = append(results, Result{
			Title:       entry.Title,
			Authors:     authors,
			Summary:     entry.Summary,
			PdfURL:      pdfURL,
			PublishedAt: entry.Published,
		})
	}

	return client.formatResults(results), nil
}

// formatResults will return a structured string with the results.
func (client *Client) formatResults(results []Result) string {
	var formattedResults strings.Builder

	for _, result := range results {
		formattedResults.WriteString(fmt.Sprintf("Title: %s\n", result.Title))
		formattedResults.WriteString(fmt.Sprintf("Authors: %s\n", strings.Join(result.Authors, ", ")))
		formattedResults.WriteString(fmt.Sprintf("Summary: %s\n", result.Summary))
		formattedResults.WriteString(fmt.Sprintf("PDF URL: %s\n", result.PdfURL))
		formattedResults.WriteString(fmt.Sprintf("Published: %s\n\n", result.PublishedAt))
	}

	return formattedResults.String()
}
