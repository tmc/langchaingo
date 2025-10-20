package llms

// ReasoningOptions configures reasoning/thinking behavior for LLMs that support it.
//
// Provider Support:
//   - Anthropic Claude 4: Comprehensive thinking support with ThinkingMode and Interleaved thinking
//   - OpenAI o1/o3: Reasoning effort controlled via Strength parameter
//   - Google Gemini 2.0+: Thinking mode support (experimental)
//
// Example usage:
//
//	// Anthropic Claude 4 with thinking
//	resp, err := llm.GenerateContent(ctx, messages,
//	    llms.WithReasoning(llms.ReasoningOptions{
//	        Mode:        llms.ThinkingModeMedium,
//	        Interleaved: true, // Think between tool calls
//	    }),
//	    llms.WithTemperature(1.0), // Required for Anthropic thinking
//	)
//
//	// OpenAI o1/o3 with reasoning effort
//	resp, err := llm.GenerateContent(ctx, messages,
//	    llms.WithReasoningStrength(0.8), // High reasoning effort
//	)
type ReasoningOptions struct {
	// Mode controls the level of reasoning/thinking (Anthropic, Google).
	// Higher modes use more tokens but provide deeper reasoning.
	Mode ThinkingMode

	// Strength controls reasoning effort for OpenAI o1/o3 models (0.0-1.0).
	// 0.0-0.33: Low effort, 0.33-0.66: Medium effort, 0.66-1.0: High effort.
	// This maps to OpenAI's "reasoning_effort" parameter.
	Strength *float64

	// BudgetTokens limits the number of tokens used for thinking (Anthropic).
	// If not set, the model determines the budget automatically.
	BudgetTokens *int

	// Interleaved enables thinking between tool calls (Anthropic beta feature).
	// When true, the model can reason about which tools to use and plan multi-step workflows.
	Interleaved bool
}

// ThinkingMode controls the level of reasoning/thinking.
type ThinkingMode string

const (
	// ThinkingModeLow provides minimal thinking/reasoning.
	// Uses fewer tokens, faster responses, basic reasoning.
	ThinkingModeLow ThinkingMode = "low"

	// ThinkingModeMedium provides balanced thinking/reasoning.
	// Moderate token usage, good balance of speed and quality.
	ThinkingModeMedium ThinkingMode = "medium"

	// ThinkingModeHigh provides deep thinking/reasoning.
	// Uses more tokens, slower responses, highest quality reasoning.
	ThinkingModeHigh ThinkingMode = "high"
)

// ReasoningUsage provides standardized tracking of reasoning/thinking tokens across providers.
type ReasoningUsage struct {
	ReasoningTokens  int
	OutputTokens     int
	ReasoningContent string
	ThinkingContent  string
	BudgetAllocated  int
	BudgetUsed       int
}

// WithReasoning configures reasoning/thinking options for the LLM.
func WithReasoning(opts ReasoningOptions) CallOption {
	return func(o *CallOptions) {
		o.Reasoning = &opts
	}
}

// WithThinkingMode is a convenience function for setting the thinking mode.
func WithThinkingMode(mode ThinkingMode) CallOption {
	return func(o *CallOptions) {
		if o.Reasoning == nil {
			o.Reasoning = &ReasoningOptions{}
		}
		o.Reasoning.Mode = mode
	}
}

// WithReasoningStrength is a convenience function for setting reasoning strength (OpenAI o1/o3).
func WithReasoningStrength(strength float64) CallOption {
	return func(o *CallOptions) {
		if o.Reasoning == nil {
			o.Reasoning = &ReasoningOptions{}
		}
		o.Reasoning.Strength = &strength
	}
}

// ExtractReasoningUsage extracts reasoning/thinking token information from GenerationInfo.
func ExtractReasoningUsage(genInfo map[string]any) *ReasoningUsage {
	if genInfo == nil {
		return nil
	}

	usage := &ReasoningUsage{}

	// Try Anthropic format
	if thinking, ok := genInfo["ThinkingContent"].(string); ok {
		usage.ThinkingContent = thinking
		usage.ReasoningContent = thinking
	}
	if tokens, ok := genInfo["ThinkingTokens"].(int); ok {
		usage.ReasoningTokens = tokens
	}
	if allocated, ok := genInfo["ThinkingBudgetAllocated"].(int); ok {
		usage.BudgetAllocated = allocated
	}
	if used, ok := genInfo["ThinkingBudgetUsed"].(int); ok {
		usage.BudgetUsed = used
	}

	// Try OpenAI format
	if reasoning, ok := genInfo["ReasoningContent"].(string); ok {
		usage.ReasoningContent = reasoning
	}
	if tokens, ok := genInfo["ReasoningTokens"].(int); ok {
		usage.ReasoningTokens = tokens
	}

	if usage.ReasoningTokens == 0 && usage.ReasoningContent == "" {
		return nil
	}

	return usage
}
