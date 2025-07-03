package prompts

// RenderOption configures template rendering behavior.
type RenderOption func(*renderConfig)

// renderConfig holds internal configuration for rendering.
type renderConfig struct {
	// enableSanitization enables HTML escaping for safety
	enableSanitization bool
}

// WithSanitization enables HTML escaping of template data values.
// This helps prevent XSS attacks when rendering templates with untrusted data.
// When enabled, HTML special characters in string values will be escaped.
func WithSanitization() RenderOption {
	return func(c *renderConfig) {
		c.enableSanitization = true
	}
}

// applyOptions applies the given options to a render configuration.
func applyOptions(opts []RenderOption) *renderConfig {
	cfg := &renderConfig{
		enableSanitization: false, // Sanitization is opt-in
	}
	for _, opt := range opts {
		opt(cfg)
	}
	return cfg
}
