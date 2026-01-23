package llms_test

import (
	"testing"

	"github.com/vxcontrol/langchaingo/llms"
)

func TestIsReasoningModel(t *testing.T) {
	tests := []struct {
		model    string
		expected bool
	}{
		// OpenAI reasoning models
		{"gpt-5", true},
		{"gpt-5-mini", true},
		{"gpt-5-preview", true},
		{"gpt-5.1", true},
		{"gpt-5.2-pro", true},
		{"gpt-oss-120b", true},
		{"gpt-oss-20b", true},
		{"o1-preview", true},
		{"o1-mini", true},
		{"o1-pro", true},
		{"o3", true},
		{"o3-mini", true},
		{"o3-2025-04-16", true},
		{"o3-deep-research", true},
		{"o3-pro", true},
		{"o4-mini", true},
		{"o4-mini-high", true},
		{"openai/gpt-5", true},
		{"openai/o1-preview", true},

		// Anthropic extended thinking models
		{"claude-3.7-sonnet", true},
		{"claude-3.7-sonnet:thinking", true},
		{"claude-opus-4", true},
		{"claude-opus-4.1", true},
		{"claude-opus-4.5", true},
		{"claude-sonnet-4", true},
		{"claude-sonnet-4.5", true},
		{"claude-haiku-4.5", true},
		{"anthropic/claude-opus-4", true},
		{"anthropic/claude-3.7-sonnet", true},

		// DeepSeek reasoning models
		{"deepseek-r1", true},
		{"deepseek-r1-0528", true},
		{"deepseek-r1-distill-llama-70b", true},
		{"deepseek-chat-v3-0324", true},
		{"deepseek-chat-v3.1", true},
		{"deepseek-v3.1-terminus", true},
		{"deepseek-v3.2", true},
		{"deepseek-v3.2-exp", true},
		{"deepseek/deepseek-r1", true},

		// Google Gemini reasoning models
		{"gemini-2.5-flash", true},
		{"gemini-2.5-flash-lite", true},
		{"gemini-2.5-pro", true},
		{"gemini-2.5-pro-preview", true},
		{"gemini-3-flash-preview", true},
		{"gemini-3-pro-preview", true},
		{"google/gemini-2.5-flash", true},

		// X-AI Grok reasoning models
		{"grok-3-mini", true},
		{"grok-3-mini-beta", true},
		{"grok-4", true},
		{"grok-4-fast", true},
		{"grok-4.1-fast", true},
		{"grok-code-fast-1", true},
		{"x-ai/grok-4", true},

		// Z-AI GLM reasoning models (Zhipu AI)
		{"glm-4.5", true},
		{"glm-4.5-air", true},
		{"glm-4.5v", true},
		{"glm-4.6", true},
		{"glm-4.6v", true},
		{"glm-4.7", true},
		{"glm-4.7-flash", true},
		{"z-ai/glm-4.5", true},

		// Qwen reasoning models
		{"qwen3-235b-a22b-thinking-2507", true},
		{"qwen3-30b-a3b-thinking-2507", true},
		{"qwen3-next-80b-a3b-thinking", true},
		{"qwen3-vl-235b-a22b-thinking", true},
		{"qwq-32b", true},
		{"qwen/qwq-32b", true},

		// Minimax reasoning models
		{"minimax-m1", true},
		{"minimax-m2", true},
		{"minimax-m2.1", true},
		{"minimax/minimax-m1", true},

		// Moonshot AI Kimi reasoning models
		{"kimi-k2-thinking", true},
		{"kimi-dev-72b", true},
		{"moonshotai/kimi-k2-thinking", true},
		{"moonshotai/kimi-dev-72b", true},

		// Perplexity reasoning models
		{"sonar-deep-research", true},
		{"sonar-pro-search", true},
		{"sonar-reasoning-pro", true},
		{"perplexity/sonar-deep-research", true},

		// Other reasoning models
		{"aion-1.0", true},
		{"aion-1.0-mini", true},
		{"olmo-3-32b-think", true},
		{"olmo-3.1-32b-think", true},
		{"nova-2-lite-v1", true},
		{"trinity-mini", true},
		{"ernie-4.5-21b-a3b-thinking", true},
		{"seed-1.6", true},
		{"cogito-v2-preview-llama-109b-moe", true},
		{"cogito-v2.1-671b", true},
		{"lfm-2.5-1.2b-thinking:free", true},
		{"deephermes-3-mistral-24b-preview", true},
		{"hermes-4-405b", true},
		{"nemotron-3-nano-30b-a3b", true},
		{"intellect-3", true},
		{"step3", true},
		{"hunyuan-a13b-instruct", true},
		{"deepseek-r1t-chimera", true},
		{"tng-r1t-chimera", true},
		{"mimo-v2-flash", true},
		{"tongyi-deepresearch-30b-a3b", true},

		// Non-reasoning models
		{"gpt-4", false},
		{"gpt-4-turbo", false},
		{"gpt-3.5-turbo", false},
		{"claude-3-sonnet", false},
		{"claude-3-opus", false},
		{"claude-3-haiku", false},
		{"claude-2", false},
		{"claude-2.1", false},
		{"llama-2", false},
		{"llama-3", false},
		{"gemini-1.5-pro", false},
		{"gemini-1.5-flash", false},
		{"gemini-pro", false},
		{"mistral-large", false},
		{"deepseek-coder", false},
		{"qwen-72b", false},
		{"grok-1", false},
		{"grok-2", false},
	}

	for _, tt := range tests {
		t.Run(tt.model, func(t *testing.T) {
			result := llms.IsReasoningModel(tt.model)
			if result != tt.expected {
				t.Errorf("IsReasoningModel(%q) = %v, want %v", tt.model, result, tt.expected)
			}
		})
	}
}
