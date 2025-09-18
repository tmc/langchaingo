package llms

import "time"

// CacheDuration represents predefined cache durations for different use cases
type CacheDuration int

const (
	CacheShort  CacheDuration = iota // ~5 minutes - for quick iterations
	CacheMedium                      // ~1 hour - for session-based work
	CacheLong                        // ~24 hours - for stable prompts
	CacheMax                         // Provider maximum - for very stable content
)

// CacheControl represents prompt caching configuration for providers that support it.
type CacheControl struct {
	// Type specifies the type of caching (provider-specific, e.g., "ephemeral").
	Type string `json:"type,omitempty"`

	// Duration specifies cache lifetime (provider-specific limits apply).
	Duration time.Duration `json:"-"`
}

// CacheControlOptions represents advanced caching configuration
type CacheControlOptions struct {
	// Duration specifies cache lifetime. If zero, uses provider defaults.
	Duration time.Duration

	// Preset uses predefined durations optimized for common use cases
	Preset CacheDuration

	// Priority hints to the provider about cache importance (0-100)
	Priority int

	// Scope defines cache sharing ("user", "session", "global")
	Scope string

	// Tags allow grouped cache invalidation
	Tags []string
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

// WithPromptCachingAdvanced provides full control over caching behavior.
// Providers will use their best effort to honor the requested configuration.
func WithPromptCachingAdvanced(cacheOpts CacheControlOptions) CallOption {
	return func(opts *CallOptions) {
		if opts.Metadata == nil {
			opts.Metadata = make(map[string]interface{})
		}
		opts.Metadata["prompt_caching"] = true
		opts.Metadata["prompt_caching_advanced"] = cacheOpts
	}
}

// WithCacheDuration is a convenience function for setting cache duration.
func WithCacheDuration(duration time.Duration) CallOption {
	return WithPromptCachingAdvanced(CacheControlOptions{Duration: duration})
}

// WithCachePreset uses predefined durations optimized for common use cases.
func WithCachePreset(preset CacheDuration) CallOption {
	return WithPromptCachingAdvanced(CacheControlOptions{Preset: preset})
}

// resolveCacheDuration converts presets to actual durations based on provider capabilities
func resolveCacheDuration(preset CacheDuration, providerMax time.Duration) time.Duration {
	switch preset {
	case CacheShort:
		return 5 * time.Minute
	case CacheMedium:
		return 1 * time.Hour
	case CacheLong:
		return 24 * time.Hour
	case CacheMax:
		if providerMax > 0 {
			return providerMax
		}
		return 24 * time.Hour // fallback
	default:
		return 0 // provider default
	}
}
