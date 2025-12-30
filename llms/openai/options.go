package openai

import "github.com/tmc/langchaingo/llms"

// WithMaxCompletionTokens sets the max_completion_tokens field for token generation.
// This is the recommended way to limit tokens with OpenAI models.
//
// Usage:
//
//	llm.GenerateContent(ctx, messages,
//	    openai.WithMaxCompletionTokens(100),
//	)
//
// Note: While llms.WithMaxTokens() still works for backward compatibility,
// WithMaxCompletionTokens is preferred for clarity when using OpenAI.
func WithMaxCompletionTokens(maxTokens int) llms.CallOption {
	return func(opts *llms.CallOptions) {
		opts.MaxTokens = maxTokens
	}
}

// WithLegacyMaxTokensField forces the use of the max_tokens field instead of max_completion_tokens.
// This is useful when connecting to older OpenAI-compatible inference servers that only
// support the max_tokens field and don't recognize max_completion_tokens.
//
// Usage:
//
//	llm.GenerateContent(ctx, messages,
//	    llms.WithMaxTokens(100),
//	    openai.WithLegacyMaxTokensField(), // Forces use of max_tokens field
//	)
func WithLegacyMaxTokensField() llms.CallOption {
	return func(opts *llms.CallOptions) {
		if opts.Metadata == nil {
			opts.Metadata = make(map[string]interface{})
		}
		opts.Metadata["openai:use_legacy_max_tokens"] = true
	}
}

// WithExtraBody allows passing custom parameters directly to the OpenAI API.
// This is useful for beta features or new parameters not yet supported by the library.
// Fields in extraBody will be merged into the JSON request body.
//
// Usage:
//
//	llm.GenerateContent(ctx, messages,
//	    openai.WithExtraBody(map[string]interface{}{
//	        "parallel_tool_calls": false,
//	    }),
//	)
func WithExtraBody(extraBody map[string]interface{}) llms.CallOption {
	return func(opts *llms.CallOptions) {
		// Only set if extraBody is not nil and not empty
		if extraBody != nil && len(extraBody) > 0 {
			if opts.Metadata == nil {
				opts.Metadata = make(map[string]interface{})
			}
			opts.Metadata["openai:extra_body"] = extraBody
		}
	}
}
