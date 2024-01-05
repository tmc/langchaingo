package ollama

type chatOptions struct {
	options
}

type ChatOption func(*chatOptions)

// WithLLMOptions Set underlying LLM options.
func WithLLMOptions(opts ...Option) ChatOption {
	return func(copts *chatOptions) {
		for _, opt := range opts {
			opt(&copts.options)
		}
	}
}
