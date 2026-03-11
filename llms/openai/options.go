package openai

import "github.com/vxcontrol/langchaingo/llms"

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
		opts.MaxTokens = &maxTokens
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
			opts.Metadata = make(map[string]any)
		}
		opts.Metadata["openai:use_legacy_max_tokens"] = true
	}
}

func isLegacyMaxTokensField(opts *llms.CallOptions) bool {
	if opts.Metadata == nil {
		return false
	}
	return opts.Metadata["openai:use_legacy_max_tokens"] == true
}

// WithExtraBody allows passing additional fields in the request body that will be sent
// to the OpenAI-compatible API. These fields will override any existing fields with the
// same name, allowing you to use provider-specific parameters not directly supported
// by the library.
//
// Example usage:
//
//	llm.GenerateContent(ctx, messages,
//	    openai.WithExtraBody(map[string]any{
//	        "enable_thinking": false,
//	        "chat_template_kwargs": map[string]any{
//	            "enable_thinking": false,
//	        },
//	    }),
//	)
func WithExtraBody(extraBody map[string]any) llms.CallOption {
	return func(opts *llms.CallOptions) {
		if opts.Metadata == nil {
			opts.Metadata = make(map[string]any)
		}
		opts.Metadata["openai:extra_body"] = extraBody
	}
}

func getExtraBody(opts *llms.CallOptions) map[string]any {
	if opts.Metadata == nil {
		return nil
	}
	if extraBody, ok := opts.Metadata["openai:extra_body"].(map[string]any); ok {
		return extraBody
	}
	return nil
}
