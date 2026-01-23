package llms

import "strings"

// IsReasoningModel returns true if the model is a reasoning/thinking model.
// This includes OpenAI o1/o3/GPT-5 series, Anthropic Claude 3.7+, DeepSeek reasoner, etc.
// For runtime checking of LLM instances, use SupportsReasoningModel instead.
func IsReasoningModel(model string) bool {
	return DefaultIsReasoningModel(model)
}

// DefaultIsReasoningModel provides the default reasoning model detection logic.
// This can be used by LLM implementations that want to extend rather than replace
// the default detection logic.
func DefaultIsReasoningModel(model string) bool {
	modelLower := strings.ToLower(model)

	// Remove provider prefix if present (e.g., "openai/", "anthropic/", "google/")
	if idx := strings.LastIndex(modelLower, "/"); idx != -1 {
		modelLower = modelLower[idx+1:]
	}

	// OpenAI reasoning models
	if strings.HasPrefix(modelLower, "gpt-5") ||
		strings.HasPrefix(modelLower, "gpt-oss-") ||
		strings.HasPrefix(modelLower, "o1-") ||
		strings.HasPrefix(modelLower, "o3") ||
		strings.HasPrefix(modelLower, "o4-mini") {
		return true
	}

	// Anthropic extended thinking models
	if strings.Contains(modelLower, "claude-3.7") ||
		strings.HasPrefix(modelLower, "claude-opus-4") ||
		strings.HasPrefix(modelLower, "claude-sonnet-4") ||
		strings.Contains(modelLower, "claude-haiku-4.5") {
		return true
	}

	// DeepSeek reasoning models
	if strings.Contains(modelLower, "deepseek-r1") ||
		strings.Contains(modelLower, "deepseek-chat-v3") ||
		strings.Contains(modelLower, "deepseek-v3.1-terminus") ||
		strings.HasPrefix(modelLower, "deepseek-v3.2") {
		return true
	}

	// Google Gemini reasoning models
	if strings.HasPrefix(modelLower, "gemini-2.5-") ||
		strings.HasPrefix(modelLower, "gemini-3-") {
		return true
	}

	// X-AI Grok reasoning models
	if strings.HasPrefix(modelLower, "grok-3-mini") ||
		strings.HasPrefix(modelLower, "grok-4") ||
		strings.Contains(modelLower, "grok-code-fast") {
		return true
	}

	// Z-AI GLM reasoning models (Zhipu AI)
	if strings.HasPrefix(modelLower, "glm-4.5") ||
		strings.HasPrefix(modelLower, "glm-4.6") ||
		strings.HasPrefix(modelLower, "glm-4.7") {
		return true
	}

	// Qwen reasoning models
	if (strings.HasPrefix(modelLower, "qwen") && strings.Contains(modelLower, "thinking")) ||
		strings.Contains(modelLower, "qwq-") {
		return true
	}

	// Minimax reasoning models
	if strings.HasPrefix(modelLower, "minimax-m") {
		return true
	}

	// Moonshot AI Kimi reasoning models
	if strings.Contains(modelLower, "kimi-") &&
		(strings.Contains(modelLower, "k2-thinking") || strings.Contains(modelLower, "dev-72b")) {
		return true
	}

	// Perplexity reasoning models
	if strings.Contains(modelLower, "sonar-") &&
		(strings.Contains(modelLower, "deep-research") ||
			strings.Contains(modelLower, "reasoning") ||
			strings.Contains(modelLower, "pro-search")) {
		return true
	}

	// Other reasoning models
	if strings.Contains(modelLower, "aion-") ||
		(strings.Contains(modelLower, "olmo-") && strings.Contains(modelLower, "-think")) ||
		strings.Contains(modelLower, "nova-2-lite") ||
		strings.Contains(modelLower, "trinity-mini") ||
		strings.Contains(modelLower, "ernie-4.5") ||
		strings.Contains(modelLower, "seed-1.6") ||
		strings.Contains(modelLower, "cogito-v2") ||
		(strings.Contains(modelLower, "lfm-") && strings.Contains(modelLower, "-thinking")) ||
		strings.Contains(modelLower, "deephermes") ||
		strings.Contains(modelLower, "hermes-4-") ||
		strings.Contains(modelLower, "nemotron") ||
		strings.Contains(modelLower, "intellect-3") ||
		strings.Contains(modelLower, "step3") ||
		strings.Contains(modelLower, "hunyuan-a13b") ||
		strings.Contains(modelLower, "chimera") ||
		strings.Contains(modelLower, "mimo-v2") ||
		strings.Contains(modelLower, "tongyi-deepresearch") {
		return true
	}

	return false
}
