package openai

const (
	tokenEnvVarName   = "OPENAI_API_KEY"  //nolint:gosec
	modelEnvVarName   = "OPENAI_MODEL"    //nolint:gosec
	baseUrlEnvVarName = "OPENAI_BASE_URL" //nolint:gosec
)

type options struct {
	token   string
	model   string
	baseUrl string
}

type Option func(*options)

// WithToken passes the OpenAI API token to the client. If not set, the token
// is read from the OPENAI_API_KEY environment variable.
func WithToken(token string) Option {
	return func(opts *options) {
		opts.token = token
	}
}

// WithModel passes the OpenAI model to the client. If not set, the model
// is read from the OPENAI_MODEL environment variable.
func WithModel(model string) Option {
	return func(opts *options) {
		opts.model = model
	}
}

// WithBaseUrl passes the OpenAI base url to the client. If not set, the base url
// is read from the OPENAI_BASE_URL environment variable. If still not set in ENV
// VAR OPENAI_BASE_URL, then the default value is https://api.openai.com is used.
func WithBaseUrl(baseUrl string) Option {
	return func(opts *options) {
		opts.baseUrl = baseUrl
	}
}
