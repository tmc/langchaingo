package keenable

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

const (
	// DefaultBaseURL is the Keenable search endpoint. Without an API key the
	// tool calls DefaultBaseURL + "/public".
	DefaultBaseURL = "https://api.keenable.ai/v1/search"
	// DefaultAppTitle identifies the caller on the keyless public endpoint,
	// which requires an X-Keenable-Title header.
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

	apiKey     string
	baseURL    string
	appTitle   string
	maxResults int
	site       string
	httpClient *http.Client
}

var _ tools.Tool = Tool{}

// Option defines a function for configuring the Keenable tool.
type Option func(*Tool)

// WithAPIKey sets the Keenable API key. Without one, the keyless public
// endpoint is used.
func WithAPIKey(apiKey string) Option {
	return func(t *Tool) { t.apiKey = strings.TrimSpace(apiKey) }
}

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(client *http.Client) Option {
	return func(t *Tool) { t.httpClient = client }
}

// WithMaxResults sets the number of results per query (1-50).
func WithMaxResults(n int) Option {
	return func(t *Tool) { t.maxResults = n }
}

// WithSite restricts results to a single domain, e.g. "go.dev".
func WithSite(site string) Option {
	return func(t *Tool) { t.site = strings.TrimSpace(site) }
}

// WithBaseURL overrides the search endpoint.
func WithBaseURL(baseURL string) Option {
	return func(t *Tool) { t.baseURL = strings.TrimRight(baseURL, "/") }
}

// WithAppTitle sets the X-Keenable-Title header sent to the keyless endpoint.
func WithAppTitle(title string) Option {
	return func(t *Tool) { t.appTitle = title }
}

// New initializes a new Keenable search tool. The API key defaults to the
// KEENABLE_API_KEY environment variable and is optional.
func New(opts ...Option) (*Tool, error) {
	tool := &Tool{
		apiKey:     strings.TrimSpace(os.Getenv("KEENABLE_API_KEY")),
		baseURL:    DefaultBaseURL,
		appTitle:   DefaultAppTitle,
		maxResults: DefaultMaxResults,
		httpClient: &http.Client{Transport: httputil.DefaultTransport},
	}
	for _, opt := range opts {
		opt(tool)
	}
	if tool.maxResults < 1 {
		tool.maxResults = 1
	}
	if tool.maxResults > maxResultsLimit {
		tool.maxResults = maxResultsLimit
	}
	return tool, nil
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

type searchResponse struct {
	Results []struct {
		Title       string `json:"title"`
		URL         string `json:"url"`
		Snippet     string `json:"snippet"`
		Description string `json:"description"`
		PublishedAt string `json:"published_at"`
	} `json:"results"`
}

func (t Tool) search(ctx context.Context, query string) (string, error) {
	body := map[string]any{"query": query, "max_results": t.maxResults}
	if t.site != "" {
		body["site"] = t.site
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("marshal keenable request: %w", err)
	}

	endpoint := t.baseURL
	if t.apiKey == "" {
		endpoint += "/public"
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("creating keenable request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if t.apiKey != "" {
		req.Header.Set("X-API-Key", t.apiKey)
	} else {
		req.Header.Set("X-Keenable-Title", t.appTitle)
	}

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("keenable request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body) // body only used for the error message
		return "", fmt.Errorf("keenable api responded with %d: %s", resp.StatusCode, string(b))
	}

	var parsed searchResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return "", fmt.Errorf("decode keenable response: %w", err)
	}

	var sb strings.Builder
	n := 0
	for _, r := range parsed.Results {
		if r.URL == "" {
			continue
		}
		if n == t.maxResults {
			break
		}
		text := r.Snippet
		if text == "" {
			text = r.Description
		}
		fmt.Fprintf(&sb, "Title: %s\n", r.Title)
		if r.PublishedAt != "" {
			fmt.Fprintf(&sb, "Published: %s\n", r.PublishedAt)
		}
		fmt.Fprintf(&sb, "Description: %s\nURL: %s\n\n", text, r.URL)
		n++
	}
	if n == 0 {
		return "", ErrNoGoodResult
	}
	return sb.String(), nil
}
