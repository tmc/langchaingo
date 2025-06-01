package serpapi

import "net/http"

type options struct {
	apiKey     string
	httpClient *http.Client
}

type Option func(*options)

// WithAPIKey passes the Serpapi API token to the client. If not set, the token
// is read from the SERPAPI_API_KEY environment variable.
func WithAPIKey(apiKey string) Option {
	return func(opts *options) {
		opts.apiKey = apiKey
	}
}

// WithHTTPClient sets a custom HTTP client for the SerpAPI requests.
func WithHTTPClient(client *http.Client) Option {
	return func(opts *options) {
		opts.httpClient = client
	}
}
