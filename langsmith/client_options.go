package langsmith

import "net/http"

type ClientOption interface {
	apply(c *Client)
}

type clientOptionFunc func(c *Client)

func (f clientOptionFunc) apply(c *Client) {
	f(c)
}

func WithAPIKey(apiKey string) ClientOption {
	return clientOptionFunc(func(c *Client) {
		c.apiKey = apiKey
	})
}

func WithAPIURL(apiURL string) ClientOption {
	return clientOptionFunc(func(c *Client) {
		c.apiURL = apiURL
	})
}

func WithWebURL(webURL string) ClientOption {
	return clientOptionFunc(func(c *Client) {
		c.webURL = &webURL
	})
}

func WithHTTPClient(httpClient *http.Client) ClientOption {
	return clientOptionFunc(func(c *Client) {
		c.httpClient = httpClient
	})
}

func WithHideInputs(hide bool) ClientOption {
	return clientOptionFunc(func(c *Client) {
		c.hideInputs = hide
	})
}

func WithHideOutputs(hide bool) ClientOption {
	return clientOptionFunc(func(c *Client) {
		c.hideOutputs = hide
	})
}
