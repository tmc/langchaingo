package openai

import "github.com/tmc/langchaingo/llms"

// WithMaxCompletionTokens specifies the maximum number of completion tokens to generate.
// This is the recommended way to limit token generation in OpenAI models and corresponds
// to the max_completion_tokens field in the OpenAI API.
//
// This is an OpenAI-specific option that should be used instead of the generic
// llms.WithMaxTokens for better control and to avoid API errors.
func WithMaxCompletionTokens(maxCompletionTokens int) llms.CallOption {
	return func(opts *llms.CallOptions) {
		// Store in MaxTokens for backward compatibility with the existing client logic,
		// but mark it specially so we know it's for max_completion_tokens
		opts.MaxTokens = maxCompletionTokens
		// Use metadata to store that this is specifically max_completion_tokens
		if opts.Metadata == nil {
			opts.Metadata = make(map[string]interface{})
		}
		opts.Metadata["_openai_use_max_completion_tokens"] = true
	}
}

// WithDeprecatedMaxTokens specifies the maximum number of tokens using the deprecated max_tokens field.
// This is provided for compatibility with OpenAI-compatible providers that still require this field.
//
// Note: OpenAI's official API will return a 400 error if both max_tokens and max_completion_tokens
// are set. Use this only when working with providers that require the deprecated field.
func WithDeprecatedMaxTokens(maxTokens int) llms.CallOption {
	return func(opts *llms.CallOptions) {
		opts.MaxTokens = maxTokens
		// Use metadata to store that this should use the deprecated max_tokens field
		if opts.Metadata == nil {
			opts.Metadata = make(map[string]interface{})
		}
		opts.Metadata["_openai_use_deprecated_max_tokens"] = true
	}
}
