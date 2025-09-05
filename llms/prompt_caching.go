package llms

import "time"

// CacheControl represents prompt caching configuration for providers that support it.
type CacheControl struct {
	// Type specifies the type of caching (provider-specific, e.g., "ephemeral").
	Type string `json:"type,omitempty"`

	// Duration specifies cache lifetime (provider-specific limits apply).
	Duration time.Duration `json:"-"`
}

// CachedContent represents content with caching instructions.
type CachedContent struct {
	ContentPart
	CacheControl *CacheControl `json:"cache_control,omitempty"`
}

func (cc CachedContent) isPart() {}

// WithCacheControl wraps content with cache control instructions.
func WithCacheControl(content ContentPart, control *CacheControl) CachedContent {
	return CachedContent{
		ContentPart:  content,
		CacheControl: control,
	}
}

// WithPromptCaching adds cache control metadata to call options.
// This is a generic option that can be used by any provider that supports caching.
// Providers should check for this metadata and handle it appropriately.
func WithPromptCaching(enabled bool) CallOption {
	return func(opts *CallOptions) {
		if opts.Metadata == nil {
			opts.Metadata = make(map[string]interface{})
		}
		opts.Metadata["prompt_caching"] = enabled
	}
}
