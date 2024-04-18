package serpapi

type options struct {
	apiKey string
}

type Option func(*options)

// WithAPIKey passes the Serpapi API token to the client. If not set, the token
// is read from the SERPAPI_API_KEY environment variable.
func WithAPIKey(apiKey string) Option {
	return func(opts *options) {
		opts.apiKey = apiKey
	}
}
