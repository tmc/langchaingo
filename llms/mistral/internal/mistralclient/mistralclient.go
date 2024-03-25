package mistralclient

import (
	"os"
	"time"

	"github.com/gage-technologies/mistral-go"
)

type clientOptions struct {
	apiKey     string
	endpoint   string
	maxRetries int
	timeout    time.Duration
}

type Option func(*clientOptions)

func WithAPIKey(apiKey string) Option {
	return func(o *clientOptions) {
		o.apiKey = apiKey
	}
}

func WithEndpoint(endpoint string) Option {
	return func(o *clientOptions) {
		o.endpoint = endpoint
	}
}

func WithMaxRetries(maxRetries int) Option {
	return func(o *clientOptions) {
		o.maxRetries = maxRetries
	}
}

func WithTimeout(timeout time.Duration) Option {
	return func(o *clientOptions) {
		o.timeout = timeout
	}
}

func NewClient(opts ...Option) *mistral.MistralClient {

	options := &clientOptions{
		apiKey:     os.Getenv("MISTRAL_API_KEY"),
		endpoint:   mistral.Endpoint,
		maxRetries: mistral.DefaultMaxRetries,
		timeout:    mistral.DefaultTimeout,
	}

	for _, opt := range opts {
		opt(options)
	}

	return mistral.NewMistralClient(options.apiKey, options.endpoint, options.maxRetries, options.timeout)
}
