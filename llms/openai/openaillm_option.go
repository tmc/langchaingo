package openai

const (
	tokenEnvVarName = "OPENAI_API_KEY" //nolint:gosec
	modelEnvVarName = "OPENAI_MODEL"   //nolint:gosec
)

type options struct {
	token string
	model string
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
