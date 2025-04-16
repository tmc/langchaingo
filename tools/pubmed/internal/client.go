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
	"time"
)

// Client defines an HTTP client for communicating with PubMed.
type Client struct {
	MaxResults int
	UserAgent  string
	BaseURL    string
}

// Result defines a search query result type.
type Result struct {
	Title     string
	Authors   []string
	Abstract  string
	PMID      string
	Published string
}

var (
	ErrNoGoodResult = errors.New("no good search results found")
	ErrAPIResponse  = errors.New("PubMed API responded with error")
)

// NewClient initializes a Client with arguments for setting a max
// results per search query and a value for the user agent header.
func NewClient(maxResults int, userAgent string) *Client {
	if maxResults == 0 {
		maxResults = 1
	}

	return &Client{
		MaxResults: maxResults,
		UserAgent:  userAgent,
		BaseURL:    "https://eutils.ncbi.nlm.nih.gov/entrez/eutils",
	}
}

func (client *Client) newRequest(ctx context.Context, queryURL string) (*http.Request, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, queryURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating PubMed request: %w", err)
	}

	if client.UserAgent != "" {
		request.Header.Add("User-Agent", client.UserAgent)
	}

	return request, nil
}

// Search performs a search query and returns
// the result as string and an error if any.
func (client *Client) Search(ctx context.Context, query string) (string, error) {
	searchResult, err := client.searchIDs(ctx, query)
	if err != nil {
		return "", err
	}

	results, err := client.fetchArticles(ctx, searchResult.WebEnv, searchResult.QueryKey)
	if err != nil {
		return "", err
	}

	return client.formatResults(results), nil
}

func (client *Client) searchIDs(ctx context.Context, query string) (*searchResult, error) {
	searchURL := fmt.Sprintf("%s/esearch.fcgi?db=pubmed&term=%s&retmax=%d&usehistory=y",
		client.BaseURL, url.QueryEscape(query), client.MaxResults)

	request, err := client.newRequest(ctx, searchURL)
	if err != nil {
		return nil, err
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("get %s error: %w", searchURL, err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, ErrAPIResponse
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	var result searchResult
	if err := xml.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("unmarshaling XML: %w", err)
	}

	if len(result.IDList.IDs) == 0 {
		return nil, ErrNoGoodResult
	}

	return &result, nil
}

func (client *Client) fetchArticles(ctx context.Context, webEnv, queryKey string) ([]Result, error) {
	fetchURL := fmt.Sprintf("%s/efetch.fcgi?db=pubmed&WebEnv=%s&query_key=%s&retmode=xml&rettype=abstract",
		client.BaseURL, webEnv, queryKey)

	request, err := client.newRequest(ctx, fetchURL)
	if err != nil {
		return nil, err
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("get %s error: %w", fetchURL, err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, ErrAPIResponse
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	var fetchResult fetchResult
	if err := xml.Unmarshal(body, &fetchResult); err != nil {
		return nil, fmt.Errorf("unmarshaling XML: %w", err)
	}

	return client.processArticles(fetchResult.Articles), nil
}

func (client *Client) processArticles(articles []article) []Result {
	results := make([]Result, len(articles))
	for i, article := range articles {
		authors := make([]string, len(article.MedlineCitation.Article.AuthorList.Authors))
		for j, author := range article.MedlineCitation.Article.AuthorList.Authors {
			authors[j] = fmt.Sprintf("%s %s", author.ForeName, author.LastName)
		}

		pubDate := client.findEarliestDate(article.PubmedData.History.PubMedPubDate)

		results[i] = Result{
			Title:     article.MedlineCitation.Article.ArticleTitle,
			Authors:   authors,
			Abstract:  article.MedlineCitation.Article.Abstract.AbstractText,
			PMID:      article.MedlineCitation.PMID,
			Published: pubDate.Format("2006-01-02"),
		}
	}
	return results
}

func (client *Client) findEarliestDate(dates []pubMedPubDate) time.Time {
	var earliestDate time.Time
	for _, date := range dates {
		if parsedDate, err := time.Parse("2006-1-2", fmt.Sprintf("%s-%s-%s", date.Year, date.Month, date.Day)); err == nil {
			if earliestDate.IsZero() || parsedDate.Before(earliestDate) {
				earliestDate = parsedDate
			}
		}
	}
	return earliestDate
}

// formatResults will return a structured string with the results.
func (client *Client) formatResults(results []Result) string {
	var formattedResults strings.Builder

	for _, result := range results {
		formattedResults.WriteString(fmt.Sprintf("Title: %s\n", result.Title))
		formattedResults.WriteString(fmt.Sprintf("Authors: %s\n", strings.Join(result.Authors, ", ")))
		formattedResults.WriteString(fmt.Sprintf("Abstract: %s\n", result.Abstract))
		formattedResults.WriteString(fmt.Sprintf("PMID: %s\n", result.PMID))
		formattedResults.WriteString(fmt.Sprintf("Published: %s\n\n", result.Published))
	}

	return formattedResults.String()
}

// searchResult defines the structure of a search query result.
type searchResult struct {
	IDList struct {
		IDs []string `xml:"Id"`
	} `xml:"IdList"`
	WebEnv   string `xml:"WebEnv"`
	QueryKey string `xml:"QueryKey"`
}

type fetchResult struct {
	Articles []article `xml:"PubmedArticle"`
}

type article struct {
	MedlineCitation struct {
		Article struct {
			ArticleTitle string `xml:"ArticleTitle"`
			Abstract     struct {
				AbstractText string `xml:"AbstractText"`
			} `xml:"Abstract"`
			AuthorList struct {
				Authors []struct {
					LastName string `xml:"LastName"`
					ForeName string `xml:"ForeName"`
					Initials string `xml:"Initials"`
				} `xml:"Author"`
			} `xml:"AuthorList"`
		} `xml:"Article"`
		PMID string `xml:"PMID"`
	} `xml:"MedlineCitation"`
	PubmedData struct {
		History struct {
			PubMedPubDate []pubMedPubDate `xml:"PubMedPubDate"`
		} `xml:"History"`
	} `xml:"PubmedData"`
}

type pubMedPubDate struct {
	Year  string `xml:"Year"`
	Month string `xml:"Month"`
	Day   string `xml:"Day"`
}
