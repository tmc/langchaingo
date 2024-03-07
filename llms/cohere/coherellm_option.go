package cohere

const (
	tokenEnvVarName   = "COHERE_API_KEY"  //nolint:gosec
	modelEnvVarName   = "COHERE_MODEL"    //nolint:gosec
	baseURLEnvVarName = "COHERE_BASE_URL" //nolint:gosec
)

type options struct {
	token   string
	model   string
	baseURL string
}

type Option func(*options)

// WithToken passes the Cohere API token to the client. If not set, the token
// is read from the COHERE_API_KEY environment variable.
func WithToken(token string) Option {
	return func(opts *options) {
		opts.token = token
	}
}

// WithModel passes the Cohere model to the client. If not set, the model
// is read from the COHERE_MODEL environment variable.
func WithModel(model string) Option {
	return func(opts *options) {
		opts.model = model
	}
}

// WithBaseURL passes the Cohere base url to the client. If not set, the base url
// is read from the COHERE_BASE_URL environment variable. If still not set in ENV
// VAR COHERE_BASE_URL, then the default value is https://api.cohere.ai is used.
func WithBaseURL(baseURL string) Option {
	return func(opts *options) {
		opts.baseURL = baseURL
	}
}
