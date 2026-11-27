package outcrawl

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/httputil"
	"github.com/tmc/langchaingo/tools"
)

const defaultBaseURL = "https://api.outcrawl.ai/v1"

// maxResponseBytes caps how much of an API response is read.
const maxResponseBytes = 10 << 20

// ErrMissingToken is returned when no API key is available.
var ErrMissingToken = errors.New("missing the outcrawl API key, set it in the OUTCRAWL_API_KEY environment variable")

// ErrAPIError is returned when the Outcrawl API responds with an error.
var ErrAPIError = errors.New("error from Outcrawl")

// ErrNoResult is returned when the API returns no usable content.
var ErrNoResult = errors.New("no results found")

type client struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

func newClient(opts []Option) (*client, *options, error) {
	o := &options{
		apiKey:          os.Getenv("OUTCRAWL_API_KEY"),
		baseURL:         defaultBaseURL,
		onlyMainContent: true,
	}
	for _, opt := range opts {
		opt(o)
	}
	if o.apiKey == "" {
		return nil, nil, ErrMissingToken
	}
	if o.httpClient == nil {
		o.httpClient = httputil.DefaultClient
	}
	return &client{
		apiKey:     o.apiKey,
		baseURL:    strings.TrimRight(o.baseURL, "/"),
		httpClient: o.httpClient,
	}, o, nil
}

func (c *client) post(ctx context.Context, path string, payload, out any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("encoding request in outcrawl: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("creating request in outcrawl: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	res, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("doing request in outcrawl: %w", err)
	}
	defer res.Body.Close()

	data, err := io.ReadAll(io.LimitReader(res.Body, maxResponseBytes))
	if err != nil {
		return fmt.Errorf("reading response in outcrawl: %w", err)
	}

	if res.StatusCode < 200 || res.StatusCode > 299 {
		var apiErr struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		}
		if json.Unmarshal(data, &apiErr) == nil && apiErr.Message != "" {
			return fmt.Errorf("%w: %d %s: %s", ErrAPIError, res.StatusCode, apiErr.Code, apiErr.Message)
		}
		return fmt.Errorf("%w: status %d", ErrAPIError, res.StatusCode)
	}

	if err := json.Unmarshal(data, out); err != nil {
		return fmt.Errorf("unmarshal data in outcrawl: %w", err)
	}
	return nil
}

// SearchTool searches the web with Outcrawl.
type SearchTool struct {
	CallbacksHandler callbacks.Handler
	client           *client
	limit            int
}

var _ tools.Tool = SearchTool{}

// NewSearch creates a new Outcrawl search tool.
func NewSearch(opts ...Option) (*SearchTool, error) {
	c, o, err := newClient(opts)
	if err != nil {
		return nil, err
	}
	return &SearchTool{client: c, limit: o.limit}, nil
}

// Name returns the name of the tool.
func (t SearchTool) Name() string {
	return "OutcrawlSearch"
}

// Description returns a description of the tool for the agent.
func (t SearchTool) Description() string {
	return `
	"A web search tool powered by Outcrawl. "
	"Useful for when you need to answer questions about current events or find pages on the internet. "
	"Input should be a search query. "
	"Returns a numbered list of results with title, URL and a short snippet."`
}

type searchRequest struct {
	Query string `json:"query"`
	Limit int    `json:"limit,omitempty"`
}

type searchResponse struct {
	Hits []struct {
		Title   string `json:"title"`
		URL     string `json:"url"`
		Snippet string `json:"snippet"`
	} `json:"hits"`
}

// Call runs the search and returns the hits as text.
func (t SearchTool) Call(ctx context.Context, input string) (string, error) {
	if t.CallbacksHandler != nil {
		t.CallbacksHandler.HandleToolStart(ctx, input)
	}

	var res searchResponse
	err := t.client.post(ctx, "/search", searchRequest{Query: input, Limit: t.limit}, &res)
	if err != nil {
		if t.CallbacksHandler != nil {
			t.CallbacksHandler.HandleToolError(ctx, err)
		}
		return "", err
	}

	if len(res.Hits) == 0 {
		return "No good Outcrawl search results were found", nil
	}

	var sb strings.Builder
	for i, hit := range res.Hits {
		if i > 0 {
			sb.WriteString("\n")
		}
		fmt.Fprintf(&sb, "%d. %s\n%s", i+1, strings.TrimSpace(hit.Title), hit.URL)
		if snippet := strings.Join(strings.Fields(hit.Snippet), " "); snippet != "" {
			sb.WriteString("\n" + snippet)
		}
		sb.WriteString("\n")
	}
	result := sb.String()

	if t.CallbacksHandler != nil {
		t.CallbacksHandler.HandleToolEnd(ctx, result)
	}
	return result, nil
}

// ScrapeTool fetches a page with Outcrawl and returns it as markdown.
type ScrapeTool struct {
	CallbacksHandler callbacks.Handler
	client           *client
	onlyMainContent  bool
}

var _ tools.Tool = ScrapeTool{}

// NewScrape creates a new Outcrawl scrape tool.
func NewScrape(opts ...Option) (*ScrapeTool, error) {
	c, o, err := newClient(opts)
	if err != nil {
		return nil, err
	}
	return &ScrapeTool{client: c, onlyMainContent: o.onlyMainContent}, nil
}

// Name returns the name of the tool.
func (t ScrapeTool) Name() string {
	return "OutcrawlScrape"
}

// Description returns a description of the tool for the agent.
func (t ScrapeTool) Description() string {
	return `
	"Fetches a web page with Outcrawl and returns its main content as markdown. "
	"Useful for when you need to read the full text of a page, including pages that render with JavaScript. "
	"Input should be a full URL starting with http:// or https://."`
}

type scrapeRequest struct {
	URL             string   `json:"url"`
	Formats         []string `json:"formats"`
	OnlyMainContent bool     `json:"onlyMainContent"`
}

type scrapeResponse struct {
	Markdown string `json:"markdown"`
}

// Call fetches the URL and returns its content as markdown.
func (t ScrapeTool) Call(ctx context.Context, input string) (string, error) {
	if t.CallbacksHandler != nil {
		t.CallbacksHandler.HandleToolStart(ctx, input)
	}

	req := scrapeRequest{
		URL:             strings.TrimSpace(input),
		Formats:         []string{"markdown"},
		OnlyMainContent: t.onlyMainContent,
	}
	var res scrapeResponse
	err := t.client.post(ctx, "/scrape", req, &res)
	if err == nil && strings.TrimSpace(res.Markdown) == "" {
		err = ErrNoResult
	}
	if err != nil {
		if t.CallbacksHandler != nil {
			t.CallbacksHandler.HandleToolError(ctx, err)
		}
		return "", err
	}

	if t.CallbacksHandler != nil {
		t.CallbacksHandler.HandleToolEnd(ctx, res.Markdown)
	}
	return res.Markdown, nil
}
