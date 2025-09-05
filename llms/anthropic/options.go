package anthropic

import (
	"time"

	"github.com/tmc/langchaingo/llms"
)

// WithPromptCaching enables Anthropic's prompt caching feature.
// This allows frequently-used prompts and system messages to be cached for improved
// performance and reduced costs.
//
// Usage:
//
//	llm.GenerateContent(ctx, messages,
//	    anthropic.WithPromptCaching(),
//	)
func WithPromptCaching() llms.CallOption {
	return func(opts *llms.CallOptions) {
		if opts.Metadata == nil {
			opts.Metadata = make(map[string]interface{})
		}
		opts.Metadata["anthropic:beta_headers"] = []string{"prompt-caching-2024-07-31"}
	}
}

// WithExtendedOutput enables 128K token output for Claude 3.7+.
// Standard models are limited to 8K tokens, but this beta feature allows
// generating much longer responses.
//
// Usage:
//
//	llm.GenerateContent(ctx, messages,
//	    llms.WithMaxTokens(50000),
//	    anthropic.WithExtendedOutput(),
//	)
func WithExtendedOutput() llms.CallOption {
	return func(opts *llms.CallOptions) {
		if opts.Metadata == nil {
			opts.Metadata = make(map[string]interface{})
		}
		// Add to existing headers if present
		if existing, ok := opts.Metadata["anthropic:beta_headers"].([]string); ok {
			opts.Metadata["anthropic:beta_headers"] = append(existing, "output-128k-2025-02-19")
		} else {
			opts.Metadata["anthropic:beta_headers"] = []string{"output-128k-2025-02-19"}
		}
	}
}

// WithInterleavedThinking enables thinking between tool calls for Claude 3.7+.
// This allows the model to use reasoning tokens to plan tool usage and interpret results.
//
// Usage:
//
//	llm.GenerateContent(ctx, messages,
//	    llms.WithTools(tools),
//	    llms.WithThinkingMode(llms.ThinkingModeMedium),
//	    anthropic.WithInterleavedThinking(),
//	)
func WithInterleavedThinking() llms.CallOption {
	return func(opts *llms.CallOptions) {
		if opts.Metadata == nil {
			opts.Metadata = make(map[string]interface{})
		}
		// Add to existing headers if present
		if existing, ok := opts.Metadata["anthropic:beta_headers"].([]string); ok {
			opts.Metadata["anthropic:beta_headers"] = append(existing, "interleaved-thinking-2025-05-14")
		} else {
			opts.Metadata["anthropic:beta_headers"] = []string{"interleaved-thinking-2025-05-14"}
		}
	}
}

// WithBetaHeader adds a custom beta header for accessing Anthropic's experimental features.
// This is useful for testing new features before dedicated support is added.
//
// Usage:
//
//	llm.GenerateContent(ctx, messages,
//	    anthropic.WithBetaHeader("new-feature-2025-01-01"),
//	)
func WithBetaHeader(header string) llms.CallOption {
	return func(opts *llms.CallOptions) {
		if opts.Metadata == nil {
			opts.Metadata = make(map[string]interface{})
		}
		// Add to existing headers if present
		if existing, ok := opts.Metadata["anthropic:beta_headers"].([]string); ok {
			opts.Metadata["anthropic:beta_headers"] = append(existing, header)
		} else {
			opts.Metadata["anthropic:beta_headers"] = []string{header}
		}
	}
}

// EphemeralCache creates a standard ephemeral cache control for Anthropic with 5-minute duration.
func EphemeralCache() *llms.CacheControl {
	return &llms.CacheControl{
		Type:     "ephemeral",
		Duration: 5 * time.Minute,
	}
}

// EphemeralCacheOneHour creates a 1-hour ephemeral cache control for Anthropic.
func EphemeralCacheOneHour() *llms.CacheControl {
	return &llms.CacheControl{
		Type:     "ephemeral",
		Duration: time.Hour,
	}
}
