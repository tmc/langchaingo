package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type MetaphorClient struct {
	apiKey      string
	RequestBody RequestBody
	options     []Options
}

type RequestBody struct {
	Query              string   `json:"query,omitempty"`
	URL                string   `json:"url,omitempty"`
	NumResults         int      `json:"numResults,omitempty"`
	IncludeDomains     []string `json:"includeDomains,omitempty"`
	ExcludeDomains     []string `json:"excludeDomains,omitempty"`
	StartCrawlDate     string   `json:"startCrawlDate,omitempty"`
	EndCrawlDate       string   `json:"endCrawlDate,omitempty"`
	StartPublishedDate string   `json:"startPublishedDate,omitempty"`
	EndPublishedDate   string   `json:"endPublishedDate,omitempty"`
	UseAutoprompt      bool     `json:"useAutoprompt,omitempty"`
	Type               string   `json:"type,omitempty"`
}

const (
	// DEFAULT REQUEST VALUES.

	// DefaultNumResults is the default number of expected results.
	DefaultNumResults = 10
	// DefaultAutoprompt if true, your query will be converted to a Metaphor query.
	// If findLinks ednpoint is used needs to be nil to omit useAutoprompt filed from RequestBody.
	DefaultAutoprompt = false

	DefaultSearchType = "neutral"

	// DEFAULT API ENDPOINT URL's.

	// DefaultSearchURL is the default search endpoint.
	DefaultSearchURL = "https://api.metaphor.systems/search"
	// DefaultContentsURL is the default contents endpoint.
	DefaultContentsURL = "https://api.metaphor.systems/contents"
	// DefaultFindLinksURL is the default find links endpoint.
	DefaultFindLinksURL = "https://api.metaphor.systems/findSimilar"
)

var (
	ErrRequestFailed          = errors.New("request failed with error")
	ErrSearchFailed           = errors.New("search failed with error")
	ErrFindSimilarLinkdFailed = errors.New("find similar links failed with error")
	ErrGetContentsFailed      = errors.New("get contents failed with error")
	ErrNoSearchResults        = errors.New("no search results were found")
	ErrNoLinksFound           = errors.New("no links were found")
	ErrNoContentExtracted     = errors.New("no content was extracted")
)

func New(apiKey string, options ...Options) (*MetaphorClient, error) {
	client := &MetaphorClient{
		apiKey:      apiKey,
		RequestBody: RequestBody{},
		options:     options,
	}

	return client, nil
}

func (client *MetaphorClient) Search(
	ctx context.Context,
	query string,
	options ...Options,
) (*SearchResponse, error) {
	client.RequestBody = RequestBody{
		Query:         query,
		NumResults:    DefaultNumResults,
		UseAutoprompt: DefaultAutoprompt,
		Type:          DefaultSearchType,
	}

	client.loadOptions(options)

	searchResults := &SearchResponse{}

	reqBytes, err := json.Marshal(client.RequestBody)
	if err != nil {
		return searchResults, fmt.Errorf("%w: %w", ErrSearchFailed, err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, DefaultSearchURL, bytes.NewBuffer(reqBytes))
	if err != nil {
		return searchResults, fmt.Errorf("%w: %w", ErrSearchFailed, err)
	}

	responseBody, err := client.runRequest(req)
	if err != nil {
		return searchResults, fmt.Errorf("%w: %w", ErrSearchFailed, err)
	}

	err = json.Unmarshal(responseBody, &searchResults)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrSearchFailed, err)
	}

	if len(searchResults.Results) == 0 {
		return searchResults, ErrNoSearchResults
	}

	return searchResults, nil
}

func (client *MetaphorClient) FindSimilar(
	ctx context.Context,
	url string,
	options ...Options,
) (*SearchResponse, error) {
	client.RequestBody = RequestBody{
		URL:           url,
		UseAutoprompt: DefaultAutoprompt,
	}

	client.loadOptions(options)

	searchResults := &SearchResponse{}

	reqBytes, err := json.Marshal(client.RequestBody)
	if err != nil {
		return searchResults, fmt.Errorf("%w: %w", ErrFindSimilarLinkdFailed, err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, DefaultFindLinksURL, bytes.NewBuffer(reqBytes))
	if err != nil {
		return searchResults, fmt.Errorf("%w: %w", ErrFindSimilarLinkdFailed, err)
	}

	responseBody, err := client.runRequest(req)
	if err != nil {
		return searchResults, fmt.Errorf("%w: %w", ErrFindSimilarLinkdFailed, err)
	}

	err = json.Unmarshal(responseBody, &searchResults)
	if err != nil {
		return searchResults, fmt.Errorf("%w: %w", ErrFindSimilarLinkdFailed, err)
	}

	if len(searchResults.Results) == 0 {
		return searchResults, ErrNoLinksFound
	}

	return searchResults, nil
}

func (client *MetaphorClient) GetContents(ctx context.Context, ids []string) (*ContentsResponse, error) {
	contentsResults := &ContentsResponse{}
	joinedIds := strings.Join(ids, "\",\"")
	url := fmt.Sprintf("%s?ids=\"%s\"", DefaultContentsURL, joinedIds)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return contentsResults, fmt.Errorf("%w: %w", ErrGetContentsFailed, err)
	}

	responseBody, err := client.runRequest(req)
	if err != nil {
		return &ContentsResponse{}, fmt.Errorf("%w: %w", ErrGetContentsFailed, err)
	}

	err = json.Unmarshal(responseBody, &contentsResults)
	if err != nil {
		return contentsResults, fmt.Errorf("%w: %w", ErrGetContentsFailed, err)
	}

	if len(contentsResults.Contents) == 0 {
		return contentsResults, ErrNoSearchResults
	}

	return contentsResults, nil
}

func (client *MetaphorClient) runRequest(req *http.Request) ([]byte, error) {
	req.Header.Add("x-api-key", client.apiKey)
	req.Header.Add("accept", "application/json")
	req.Header.Add("content-type", "application/json")

	// trunk-ignore(gokart/CWE-918:-Server-Side-Request-Forgery)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		errorResponse := &ErrorResponse{}
		err := json.Unmarshal(body, &errorResponse)
		if err != nil {
			return nil, err
		}
		errorTxt := errorResponse.Text

		return nil, fmt.Errorf("%w: %s", ErrRequestFailed, errorTxt)
	}

	return body, nil
}

func (client *MetaphorClient) loadOptions(options []Options) {
	if len(options) > 0 {
		client.options = options
	}

	for _, option := range client.options {
		option(client)
	}
}
