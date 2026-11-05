package keenable

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	sdk "github.com/keenableai/keenable-sdk/go"
	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/httputil"
	"github.com/tmc/langchaingo/tools"
)

const (
	// DefaultBaseURL is the Keenable API host. The SDK appends /v1/search or
	// /v1/search/public when no API key is configured.
	DefaultBaseURL = sdk.DefaultBaseURL
	// DefaultAppTitle identifies the caller on the keyless public endpoint,
	// sent as X-Keenable-Title (SDK ClientSource).
	DefaultAppTitle = "langchaingo"
	// DefaultMaxResults is the number of results returned per query.
	DefaultMaxResults = 5
	// maxResultsLimit is the largest max_results the API accepts.
	maxResultsLimit = 50
)

// ErrNoGoodResult is returned by the client when the search returns no results.
var ErrNoGoodResult = errors.New("no good search results found")

// Tool defines a tool implementation for the Keenable web search API.
type Tool struct {
	CallbacksHandler callbacks.Handler

	client     *sdk.Client
	maxResults int
	site       string
}

var _ tools.Tool = Tool{}

// Option defines a function for configuring the Keenable tool.
type Option func(*toolConfig)

type toolConfig struct {
	withAPIKey bool
	apiKey     string
	withBase   bool
	baseURL    string
	appTitle   string
	maxResults int
	site       string
	httpClient *http.Client
}

// WithAPIKey sets the Keenable API key. Without one, the keyless public
// endpoint is used. Pass an empty string to force the keyless tier even when
// KEENABLE_API_KEY is set.
func WithAPIKey(apiKey string) Option {
	return func(c *toolConfig) {
		c.withAPIKey = true
		c.apiKey = strings.TrimSpace(apiKey)
	}
}

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(client *http.Client) Option {
	return func(c *toolConfig) { c.httpClient = client }
}

// WithMaxResults sets the number of results per query (1-50).
func WithMaxResults(n int) Option {
	return func(c *toolConfig) { c.maxResults = n }
}

// WithSite restricts results to a single domain, e.g. "go.dev".
func WithSite(site string) Option {
	return func(c *toolConfig) { c.site = strings.TrimSpace(site) }
}

// WithBaseURL overrides the Keenable API host (https://api.keenable.ai).
// A legacy value ending in /v1/search is normalized to the host root.
func WithBaseURL(baseURL string) Option {
	return func(c *toolConfig) {
		c.withBase = true
		c.baseURL = normalizeBaseURL(baseURL)
	}
}

// WithAppTitle sets the X-Keenable-Title header (SDK ClientSource) sent with
// every request, including keyed calls.
func WithAppTitle(title string) Option {
	return func(c *toolConfig) { c.appTitle = title }
}

// New initializes a new Keenable search tool. The API key defaults to the
// KEENABLE_API_KEY environment variable and is optional.
func New(opts ...Option) (*Tool, error) {
	cfg := &toolConfig{
		appTitle:   DefaultAppTitle,
		maxResults: DefaultMaxResults,
		httpClient: &http.Client{Transport: httputil.DefaultTransport},
	}
	for _, opt := range opts {
		opt(cfg)
	}
	if cfg.maxResults < 1 {
		cfg.maxResults = 1
	}
	if cfg.maxResults > maxResultsLimit {
		cfg.maxResults = maxResultsLimit
	}

	sdkOpts := &sdk.Options{
		ClientSource: cfg.appTitle,
		HTTPClient:   cfg.httpClient,
	}
	if cfg.withBase {
		base := cfg.baseURL
		sdkOpts.BaseURL = &base
	}
	if cfg.withAPIKey {
		key := cfg.apiKey
		sdkOpts.APIKey = &key
	}

	client, err := sdk.New(sdkOpts)
	if err != nil {
		return nil, err
	}

	return &Tool{
		client:     client,
		maxResults: cfg.maxResults,
		site:       cfg.site,
	}, nil
}

// Name returns a name for the tool.
func (t Tool) Name() string {
	return "Keenable Search"
}

// Description returns a description for the tool.
func (t Tool) Description() string {
	return `
	"A wrapper around Keenable web search."
	"Useful for when you need to answer questions about current events or find web pages."
	"Returns titles, URLs and long page excerpts. Works without an API key."
	"Input should be a search query."`
}

// Call performs the search and returns the formatted results.
func (t Tool) Call(ctx context.Context, input string) (string, error) {
	if t.CallbacksHandler != nil {
		t.CallbacksHandler.HandleToolStart(ctx, input)
	}

	result, err := t.search(ctx, input)
	if err != nil {
		if errors.Is(err, ErrNoGoodResult) {
			return "No good Keenable Search Results was found", nil
		}
		if t.CallbacksHandler != nil {
			t.CallbacksHandler.HandleToolError(ctx, err)
		}
		return "", err
	}

	if t.CallbacksHandler != nil {
		t.CallbacksHandler.HandleToolEnd(ctx, result)
	}
	return result, nil
}

func (t Tool) search(ctx context.Context, query string) (string, error) {
	searchOpts := &sdk.SearchOptions{MaxResults: t.maxResults}
	if t.site != "" {
		searchOpts.Site = t.site
	}

	resp, err := t.client.Search(ctx, query, searchOpts)
	if err != nil {
		return "", err
	}
	return formatSearchResults(resp.Results, t.maxResults)
}

func formatSearchResults(results []sdk.SearchResult, maxResults int) (string, error) {
	var sb strings.Builder
	n := 0
	for _, r := range results {
		if r.URL == "" {
			continue
		}
		if n == maxResults {
			break
		}
		text := r.Snippet
		if text == "" && r.Description != nil {
			text = *r.Description
		}
		fmt.Fprintf(&sb, "Title: %s\n", r.Title)
		if r.PublishedAt != nil && *r.PublishedAt != "" {
			fmt.Fprintf(&sb, "Published: %s\n", *r.PublishedAt)
		}
		fmt.Fprintf(&sb, "Description: %s\nURL: %s\n\n", text, r.URL)
		n++
	}
	if n == 0 {
		return "", ErrNoGoodResult
	}
	return sb.String(), nil
}

// normalizeBaseURL maps legacy tool base URLs (…/v1/search) to the SDK host root.
func normalizeBaseURL(baseURL string) string {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	return strings.TrimSuffix(baseURL, "/v1/search")
}
