package cloudflare

import (
	"log"
	"net/http"
	"net/url"
)

type options struct {
	cloudflareAccountID string
	cloudflareServerURL *url.URL
	cloudflareToken     string
	httpClient          *http.Client
	model               string
	embeddingModel      string
	system              string
}

type Option func(*options)

// WithModel Set the model to use.
func WithModel(model string) Option {
	return func(opts *options) {
		opts.model = model
	}
}

// WithSystemPrompt Set the system prompt. This is only valid if
// WithCustomTemplate is not set and the cloudflare model use
// .System in its model template OR if WithCustomTemplate
// is set using {{.System}}.
func WithSystemPrompt(p string) Option {
	return func(opts *options) {
		opts.system = p
	}
}

// WithAccountID Set the Account Id of the cloudflare account to use.
func WithAccountID(accountID string) Option {
	return func(opts *options) {
		opts.cloudflareAccountID = accountID
	}
}

// WithServerURL Set the URL of the cloudflare Workers AI service.
func WithServerURL(rawURL string) Option {
	return func(opts *options) {
		var err error
		opts.cloudflareServerURL, err = url.Parse(rawURL)
		if err != nil {
			log.Fatal(err)
		}
	}
}

// WithCloudflareServerURL Set the URL of the cloudflare Workers AI service.
func WithCloudflareServerURL(serverURL *url.URL) Option {
	return func(opts *options) {
		opts.cloudflareServerURL = serverURL
	}
}

// WithToken Set the token to use.
func WithToken(token string) Option {
	return func(opts *options) {
		opts.cloudflareToken = token
	}
}

func WithEmbeddingModel(model string) Option {
	return func(opts *options) {
		opts.embeddingModel = model
	}
}

// WithHTTPClient Set custom http client.
func WithHTTPClient(client *http.Client) Option {
	return func(opts *options) {
		opts.httpClient = client
	}
}
