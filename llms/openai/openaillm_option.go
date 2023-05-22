package openai

import (
	"os"
	"sync"
)

const (
	tokenEnvVarName = "OPENAI_API_KEY" //nolint:gosec
	modelEnvVarName = "OPENAI_MODEL"   //nolint:gosec
)

var (
	// nolint: gochecknoglobals
	initOptions sync.Once

	// nolint: gochecknoglobals
	defaultOptions *options
)

type options struct {
	token string
	model string
}

type Option func(*options)

// initOpts initializes defaultOptions with the environment variables.
func initOpts() {
	defaultOptions = &options{
		token: os.Getenv(tokenEnvVarName),
		model: os.Getenv(modelEnvVarName),
	}
}

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
