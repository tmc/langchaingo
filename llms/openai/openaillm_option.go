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

func WithToken(token string) Option {
	return func(opts *options) {
		opts.token = token
	}
}

func WithModel(model string) Option {
	return func(opts *options) {
		opts.model = model
	}
}
