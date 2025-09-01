package llms

import "time"

// CacheControl represents prompt caching configuration for providers that support it.
type CacheControl struct {
	// Type specifies the type of caching. For Anthropic: "ephemeral"
	Type string `json:"type,omitempty"`
	
	// Duration specifies cache lifetime. Anthropic supports 5min (default) or 1hour
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

// EphemeralCache creates a standard ephemeral cache control for Anthropic.
func EphemeralCache() *CacheControl {
	return &CacheControl{
		Type:     "ephemeral",
		Duration: 5 * time.Minute,
	}
}

// EphemeralCacheOneHour creates a 1-hour ephemeral cache control for Anthropic.
func EphemeralCacheOneHour() *CacheControl {
	return &CacheControl{
		Type:     "ephemeral", 
		Duration: time.Hour,
	}
}

// WithPromptCaching adds cache control metadata to call options.
func WithPromptCaching(enabled bool) CallOption {
	return func(opts *CallOptions) {
		if opts.Metadata == nil {
			opts.Metadata = make(map[string]interface{})
		}
		opts.Metadata["prompt_caching"] = enabled
	}
}

// WithAnthropicCachingHeaders adds required headers for Anthropic prompt caching.
func WithAnthropicCachingHeaders() CallOption {
	return func(opts *CallOptions) {
		if opts.Metadata == nil {
			opts.Metadata = make(map[string]interface{})
		}
		opts.Metadata["anthropic:beta_headers"] = []string{"prompt-caching-2024-07-31"}
	}
}