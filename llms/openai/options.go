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

// WithEnableThinking enables Qwen's deep thinking mode via OpenAI-compatible API.
// Must be used with streaming — Qwen requires enable_thinking=false for non-streaming calls.
//
// Usage:
//
//	llm.GenerateContent(ctx, messages,
//	    openai.WithEnableThinking(true),
//	    llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
//	        fmt.Print(string(chunk))
//	        return nil
//	    }),
//	)
func WithEnableThinking(enabled bool) llms.CallOption {
	return func(opts *llms.CallOptions) {
		if opts.Metadata == nil {
			opts.Metadata = make(map[string]interface{})
		}
		opts.Metadata["qwen:enable_thinking"] = enabled
	}
}

func WithEnableDeepSeekThinking(data map[string]any) llms.CallOption {
	return func(opts *llms.CallOptions) {
		if opts.Metadata == nil {
			opts.Metadata = make(map[string]interface{})
		}
		opts.Metadata["deepseek:enable_thinking"] = data
	}
}

// WithThinkingBudget limits thinking tokens for Qwen models (Qwen3+).
//
// Usage:
//
//	llm.GenerateContent(ctx, messages,
//	    openai.WithEnableThinking(true),
//	    openai.WithThinkingBudget(1000),
//	)
func WithThinkingBudget(budget int) llms.CallOption {
	return func(opts *llms.CallOptions) {
		if opts.Metadata == nil {
			opts.Metadata = make(map[string]interface{})
		}
		opts.Metadata["qwen:thinking_budget"] = budget
	}
}
