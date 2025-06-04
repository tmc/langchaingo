package ollama

import (
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/ollama/ollama/api"
)

type options struct {
	ollamaServerURL     *url.URL
	httpClient          *http.Client
	model               string
	ollamaOptions       api.Options
	customModelTemplate string
	system              string
	format              string
	keepAlive           *time.Duration
}

type Option func(*options)

// WithModel Set the model to use.
func WithModel(model string) Option {
	return func(opts *options) {
		opts.model = model
	}
}

// WithFormat Sets the Ollama output format (currently Ollama only supports "json").
func WithFormat(format string) Option {
	return func(opts *options) {
		opts.format = format
	}
}

// WithKeepAlive controls how long the model will stay loaded into memory following the request (default: 5m)
// only supported by ollama v0.1.23 and later
//
//	If set to a positive duration (e.g. 20m, 1h or 30), the model will stay loaded for the provided duration
//	If set to a negative duration (e.g. -1), the model will stay loaded indefinitely
//	If set to 0, the model will be unloaded immediately once finished
//	If not set, the model will stay loaded for 5 minutes by default
func WithKeepAlive(keepAlive string) Option {
	return func(opts *options) {
		ka, err := time.ParseDuration(keepAlive)
		if err != nil {
			log.Fatal(err)
		}
		opts.keepAlive = &ka
	}
}

// WithSystemPrompt Set the system prompt. This is only valid if
// WithCustomTemplate is not set and the ollama model use
// .System in its model template OR if WithCustomTemplate
// is set using {{.System}}.
func WithSystemPrompt(p string) Option {
	return func(opts *options) {
		opts.system = p
	}
}

// WithCustomTemplate To override the templating done on Ollama model side.
func WithCustomTemplate(template string) Option {
	return func(opts *options) {
		opts.customModelTemplate = template
	}
}

// WithServerURL Set the URL of the ollama instance to use.
func WithServerURL(rawURL string) Option {
	return func(opts *options) {
		var err error
		opts.ollamaServerURL, err = url.Parse(rawURL)
		if err != nil {
			log.Fatal(err)
		}
	}
}

// WithHTTPClient Set custom http client.
func WithHTTPClient(client *http.Client) Option {
	return func(opts *options) {
		opts.httpClient = client
	}
}

// WithOllamaOptions Set the ollama specific options.
func WithOllamaOptions(ollamaOpts api.Options) Option {
	return func(opts *options) {
		opts.ollamaOptions = ollamaOpts
	}
}
