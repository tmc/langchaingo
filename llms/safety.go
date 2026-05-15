package llms

// WithSafetyConfig adds safety configuration to call options.
func WithSafetyConfig(config map[string]any) CallOption {
	return func(opts *CallOptions) {
		if config != nil {
			opts.SafetyConfig = config
		}
	}
}
