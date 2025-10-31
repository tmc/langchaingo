package yzma

type options struct {
	model  string
	system string
}

type Option func(*options)

// WithModel sets the model to use.
func WithModel(model string) Option {
	return func(opts *options) {
		opts.model = model
	}
}

// WithSystemPrompt sets the system prompt.
func WithSystemPrompt(p string) Option {
	return func(opts *options) {
		opts.system = p
	}
}
