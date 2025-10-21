package llms

import "strings"

// ThinkingMode represents different thinking/reasoning modes for LLMs.
type ThinkingMode string

const (
	// ThinkingModeNone disables thinking/reasoning.
	ThinkingModeNone ThinkingMode = "none"

	// ThinkingModeLow allocates minimal tokens for thinking (~20% of max tokens).
	ThinkingModeLow ThinkingMode = "low"

	// ThinkingModeMedium allocates moderate tokens for thinking (~50% of max tokens).
	ThinkingModeMedium ThinkingMode = "medium"

	// ThinkingModeHigh allocates maximum tokens for thinking (~80% of max tokens).
	ThinkingModeHigh ThinkingMode = "high"

	// ThinkingModeAuto lets the model decide how much thinking is needed.
	ThinkingModeAuto ThinkingMode = "auto"
)

// ThinkingConfig configures thinking/reasoning behavior for models that support it.
type ThinkingConfig struct {
	// Mode specifies the thinking mode (none, low, medium, high, auto).
	Mode ThinkingMode `json:"mode,omitempty"`

	// BudgetTokens sets explicit token budget for thinking.
	// Providers may have different minimum and maximum limits.
	BudgetTokens int `json:"budget_tokens,omitempty"`

	// ReturnThinking controls whether thinking/reasoning is included in response.
	// Provider support and behavior varies.
	ReturnThinking bool `json:"return_thinking,omitempty"`

	// StreamThinking enables streaming of thinking tokens as they're generated.
	// Not all providers support this feature.
	StreamThinking bool `json:"stream_thinking,omitempty"`

	// InterleaveThinking enables thinking between tool calls.
	// Provider support varies.
	InterleaveThinking bool `json:"interleave_thinking,omitempty"`
}

// DefaultThinkingConfig returns a sensible default thinking configuration.
func DefaultThinkingConfig() *ThinkingConfig {
	return &ThinkingConfig{
		Mode:           ThinkingModeAuto,
		ReturnThinking: false,
		StreamThinking: false,
	}
}

// WithThinking adds thinking configuration to call options.
func WithThinking(config *ThinkingConfig) CallOption {
	return func(opts *CallOptions) {
		if opts.Metadata == nil {
			opts.Metadata = make(map[string]interface{})
		}
		opts.Metadata["thinking_config"] = config
	}
}

// getOrCreateThinkingConfig is a helper to get or create thinking config from metadata.
func getOrCreateThinkingConfig(opts *CallOptions) *ThinkingConfig {
	if opts.Metadata == nil {
		opts.Metadata = make(map[string]interface{})
	}

	if existing, ok := opts.Metadata["thinking_config"].(*ThinkingConfig); ok {
		return existing
	}

	config := DefaultThinkingConfig()
	opts.Metadata["thinking_config"] = config
	return config
}

// GetThinkingConfig safely retrieves thinking config from call options.
// Returns nil if no thinking config is present.
func GetThinkingConfig(opts *CallOptions) *ThinkingConfig {
	if opts == nil || opts.Metadata == nil {
		return nil
	}

	config, _ := opts.Metadata["thinking_config"].(*ThinkingConfig)
	return config
}

// WithThinkingMode sets the thinking mode for the request.
func WithThinkingMode(mode ThinkingMode) CallOption {
	return func(opts *CallOptions) {
		config := getOrCreateThinkingConfig(opts)
		config.Mode = mode
	}
}

// WithThinkingBudget sets explicit token budget for thinking.
func WithThinkingBudget(tokens int) CallOption {
	return func(opts *CallOptions) {
		config := getOrCreateThinkingConfig(opts)
		config.BudgetTokens = tokens
	}
}

// WithReturnThinking enables returning thinking/reasoning in the response.
func WithReturnThinking(enabled bool) CallOption {
	return func(opts *CallOptions) {
		config := getOrCreateThinkingConfig(opts)
		config.ReturnThinking = enabled
	}
}

// WithStreamThinking enables streaming of thinking tokens.
func WithStreamThinking(enabled bool) CallOption {
	return func(opts *CallOptions) {
		config := getOrCreateThinkingConfig(opts)
		config.StreamThinking = enabled
	}
}

// WithInterleaveThinking enables interleaved thinking between tool calls.
// Provider support varies - check your provider's documentation.
func WithInterleaveThinking(enabled bool) CallOption {
	return func(opts *CallOptions) {
		config := getOrCreateThinkingConfig(opts)
		config.InterleaveThinking = enabled
	}
}

// Note: ReasoningModel interface is defined in llms.go

// IsReasoningModel returns true if the model is a reasoning/thinking model.
// This includes OpenAI o1/o3/GPT-5 series, Anthropic Claude 3.7+, DeepSeek reasoner, etc.
// For runtime checking of LLM instances, use SupportsReasoningModel instead.
func IsReasoningModel(model string) bool {
	return DefaultIsReasoningModel(model)
}

// SupportsReasoningModel checks if an LLM instance supports reasoning tokens.
// This first checks if the LLM implements the ReasoningModel interface.
// If not, it falls back to checking the model string if available.
func SupportsReasoningModel(llm interface{}) bool {
	// Check if the LLM implements ReasoningModel
	if reasoner, ok := llm.(ReasoningModel); ok {
		return reasoner.SupportsReasoning()
	}

	// Fallback: check if we can extract a model string somehow
	// This is a best-effort approach for backwards compatibility
	return false
}

// DefaultIsReasoningModel provides the default reasoning model detection logic.
// This can be used by LLM implementations that want to extend rather than replace
// the default detection logic.
func DefaultIsReasoningModel(model string) bool {
	modelLower := strings.ToLower(model)

	// OpenAI reasoning models
	if strings.HasPrefix(modelLower, "gpt-5") ||
		strings.HasPrefix(modelLower, "o1-") ||
		strings.HasPrefix(modelLower, "o3-") ||
		strings.Contains(modelLower, "o1-preview") ||
		strings.Contains(modelLower, "o1-mini") ||
		strings.Contains(modelLower, "o3-mini") ||
		strings.Contains(modelLower, "o4-mini") {
		return true
	}

	// Anthropic extended thinking models
	if strings.Contains(modelLower, "claude-3-7") ||
		strings.Contains(modelLower, "claude-3.7") ||
		strings.Contains(modelLower, "claude-4") ||
		strings.Contains(modelLower, "claude-opus-4") ||
		strings.Contains(modelLower, "claude-sonnet-4") {
		return true
	}

	// DeepSeek reasoner
	if strings.Contains(modelLower, "deepseek-reasoner") ||
		strings.Contains(modelLower, "deepseek-r1") {
		return true
	}

	// Grok reasoning models
	if strings.Contains(modelLower, "grok") && strings.Contains(modelLower, "reasoning") {
		return true
	}

	return false
}

// CalculateThinkingBudget calculates the token budget based on mode and max tokens.
func CalculateThinkingBudget(mode ThinkingMode, maxTokens int) int {
	switch mode {
	case ThinkingModeLow:
		return maxTokens * 20 / 100 // 20%
	case ThinkingModeMedium:
		return maxTokens * 50 / 100 // 50%
	case ThinkingModeHigh:
		return maxTokens * 80 / 100 // 80%
	case ThinkingModeAuto:
		// Let the model decide
		return 0
	default:
		return 0
	}
}

// ThinkingTokenUsage represents token usage specific to thinking/reasoning.
type ThinkingTokenUsage struct {
	// ThinkingTokens is the total number of thinking/reasoning tokens used.
	ThinkingTokens int `json:"thinking_tokens,omitempty"`

	// ThinkingInputTokens is the number of input tokens used for thinking.
	ThinkingInputTokens int `json:"thinking_input_tokens,omitempty"`

	// ThinkingOutputTokens is the number of output tokens from thinking.
	ThinkingOutputTokens int `json:"thinking_output_tokens,omitempty"`

	// ThinkingCachedTokens is the number of cached thinking tokens (if applicable).
	ThinkingCachedTokens int `json:"thinking_cached_tokens,omitempty"`

	// ThinkingBudgetUsed is the actual budget used vs allocated.
	ThinkingBudgetUsed int `json:"thinking_budget_used,omitempty"`

	// ThinkingBudgetAllocated is the budget that was allocated.
	ThinkingBudgetAllocated int `json:"thinking_budget_allocated,omitempty"`
}

// ExtractThinkingTokens extracts thinking token information from generation info.
func ExtractThinkingTokens(generationInfo map[string]any) *ThinkingTokenUsage {
	if generationInfo == nil {
		return nil
	}

	usage := &ThinkingTokenUsage{}

	// OpenAI-style reasoning tokens
	if v, ok := generationInfo["ReasoningTokens"].(int); ok {
		usage.ThinkingTokens = v
	}
	if v, ok := generationInfo["CompletionReasoningTokens"].(int); ok {
		usage.ThinkingOutputTokens = v
	}

	// Anthropic-style thinking tokens (would be in extended thinking mode)
	if v, ok := generationInfo["ThinkingTokens"].(int); ok {
		usage.ThinkingTokens = v
	}
	if v, ok := generationInfo["ThinkingInputTokens"].(int); ok {
		usage.ThinkingInputTokens = v
	}
	if v, ok := generationInfo["ThinkingOutputTokens"].(int); ok {
		usage.ThinkingOutputTokens = v
	}

	// Budget information
	if v, ok := generationInfo["ThinkingBudgetUsed"].(int); ok {
		usage.ThinkingBudgetUsed = v
	}
	if v, ok := generationInfo["ThinkingBudgetAllocated"].(int); ok {
		usage.ThinkingBudgetAllocated = v
	}

	return usage
}
