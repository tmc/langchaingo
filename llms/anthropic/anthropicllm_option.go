package anthropic

const (
	tokenEnvVarName = "ANTHROPIC_API_KEY" //nolint:gosec
)

type options struct {
	token string
	model string
}

type Option func(*options)

// WithToken passes the Anthropic API token to the client. If not set, the token
// is read from the ANTHROPIC_API_KEY environment variable.
func WithToken(token string) Option {
	return func(opts *options) {
		opts.token = token
	}
}

// WithModel passes the Anthropic model to the client.
func WithModel(model string) Option {
	return func(opts *options) {
		opts.model = model
	}
}
