package outcrawl

import "net/http"

type options struct {
	apiKey          string
	baseURL         string
	httpClient      *http.Client
	limit           int
	onlyMainContent bool
}

// Option configures an Outcrawl tool.
type Option func(*options)

// WithAPIKey sets the Outcrawl API key. If not set, the key is read from the
// OUTCRAWL_API_KEY environment variable.
func WithAPIKey(apiKey string) Option {
	return func(opts *options) {
		opts.apiKey = apiKey
	}
}

// WithBaseURL overrides the API base URL (default https://api.outcrawl.ai/v1).
func WithBaseURL(baseURL string) Option {
	return func(opts *options) {
		opts.baseURL = baseURL
	}
}

// WithHTTPClient sets a custom HTTP client for the Outcrawl requests.
func WithHTTPClient(client *http.Client) Option {
	return func(opts *options) {
		opts.httpClient = client
	}
}

// WithLimit sets the maximum number of search hits to return. It only applies
// to the search tool. A value of zero or less uses the API default.
func WithLimit(limit int) Option {
	return func(opts *options) {
		opts.limit = limit
	}
}

// WithOnlyMainContent controls whether the scrape tool strips navigation,
// footers and similar page chrome. It only applies to the scrape tool and
// defaults to true.
func WithOnlyMainContent(onlyMainContent bool) Option {
	return func(opts *options) {
		opts.onlyMainContent = onlyMainContent
	}
}
