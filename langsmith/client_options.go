package langsmith

import "net/http"

type ClientOption interface {
	apply(c *Client)
}

type clientOptionFunc func(c *Client)

func (f clientOptionFunc) apply(c *Client) {
	f(c)
}

// WithAPIKey sets the API key for authentication.
func WithAPIKey(apiKey string) ClientOption {
	return clientOptionFunc(func(c *Client) {
		c.apiKey = apiKey
	})
}

// WithAPIURL sets the API URL.
func WithAPIURL(apiURL string) ClientOption {
	return clientOptionFunc(func(c *Client) {
		c.apiURL = apiURL
	})
}

// WithWebURL sets the web URL.
func WithWebURL(webURL string) ClientOption {
	return clientOptionFunc(func(c *Client) {
		c.webURL = &webURL
	})
}

// WithHTTPClient sets the HTTP client.
func WithHTTPClient(httpClient *http.Client) ClientOption {
	return clientOptionFunc(func(c *Client) {
		c.httpClient = httpClient
	})
}

// WithHideInputs sets whether to hide input instrumentation
func WithHideInputs(hide bool) ClientOption {
	return clientOptionFunc(func(c *Client) {
		c.hideInputs = hide
	})
}

// WithHideOutputs sets whether to hide output instrumentation
func WithHideOutputs(hide bool) ClientOption {
	return clientOptionFunc(func(c *Client) {
		c.hideOutputs = hide
	})
}

// WithClientLogger sets the logger.
func WithClientLogger(logger LeveledLogger) ClientOption {
	return clientOptionFunc(func(c *Client) {
		c.logger = logger
	})
}
