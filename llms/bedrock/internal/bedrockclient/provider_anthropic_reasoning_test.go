package bedrockclient

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vxcontrol/langchaingo/llms"
)

func marshalAnthropicInput(t *testing.T, input anthropicTextGenerationInput) map[string]any {
	t.Helper()
	raw, err := json.Marshal(input)
	require.NoError(t, err)
	var fields map[string]any
	require.NoError(t, json.Unmarshal(raw, &fields))
	return fields
}

func TestApplyAnthropicReasoning_Adaptive(t *testing.T) {
	t.Parallel()

	input := anthropicTextGenerationInput{
		MaxTokens:   2048,
		Temperature: 0.8,
		TopP:        0.9,
		TopK:        40,
	}
	applyAnthropicReasoning(&input,
		&llms.ReasoningConfig{Effort: llms.ReasoningXHigh, Adaptive: true},
		"anthropic.claude-opus-4-7-v1:0", 2048)

	fields := marshalAnthropicInput(t, input)

	thinking, _ := fields["thinking"].(map[string]any)
	assert.Equal(t, "adaptive", thinking["type"])
	assert.Equal(t, "summarized", thinking["display"])
	_, hasBudget := thinking["budget_tokens"]
	assert.False(t, hasBudget, "adaptive thinking must not carry a token budget")

	outputConfig, _ := fields["output_config"].(map[string]any)
	assert.Equal(t, "xhigh", outputConfig["effort"])

	for _, key := range []string{"temperature", "top_p", "top_k"} {
		_, has := fields[key]
		assert.False(t, has, "adaptive must omit %s", key)
	}
}

func TestApplyAnthropicReasoning_AdaptiveBypassesVersionGate(t *testing.T) {
	t.Parallel()

	// A model family newer than the budget allowlist knows must still carry adaptive.
	input := anthropicTextGenerationInput{MaxTokens: 2048}
	applyAnthropicReasoning(&input,
		&llms.ReasoningConfig{Effort: llms.ReasoningHigh, Adaptive: true},
		"us.anthropic.claude-sonnet-5-v1:0", 2048)

	require.NotNil(t, input.Thinking)
	assert.Equal(t, "adaptive", input.Thinking.Type)
	require.NotNil(t, input.OutputConfig)
	assert.Equal(t, "high", input.OutputConfig.Effort)
}

func TestApplyAnthropicReasoning_AdaptiveEmptyEffortDefaultsToHigh(t *testing.T) {
	t.Parallel()

	input := anthropicTextGenerationInput{MaxTokens: 2048}
	applyAnthropicReasoning(&input,
		&llms.ReasoningConfig{Adaptive: true},
		"anthropic.claude-opus-4-7-v1:0", 2048)

	require.NotNil(t, input.OutputConfig)
	assert.Equal(t, "high", input.OutputConfig.Effort)
}

func TestApplyAnthropicReasoning_Budget(t *testing.T) {
	t.Parallel()

	input := anthropicTextGenerationInput{
		MaxTokens:   8000,
		Temperature: 0.8,
	}
	applyAnthropicReasoning(&input,
		&llms.ReasoningConfig{Effort: llms.ReasoningMedium},
		"anthropic.claude-opus-4-6-v1:0", 8000)

	fields := marshalAnthropicInput(t, input)

	thinking, _ := fields["thinking"].(map[string]any)
	assert.Equal(t, "enabled", thinking["type"])
	assert.EqualValues(t, 2666, thinking["budget_tokens"]) // medium => max(8000/3, 2048)
	_, hasDisplay := thinking["display"]
	assert.False(t, hasDisplay, "budget thinking has no display field")

	_, hasOutputConfig := fields["output_config"]
	assert.False(t, hasOutputConfig, "budget thinking has no output_config")

	assert.EqualValues(t, 0.8, fields["temperature"], "the refactor must not alter budget sampling params")
}

func TestApplyAnthropicReasoning_BudgetKeepsVersionGate(t *testing.T) {
	t.Parallel()

	input := anthropicTextGenerationInput{MaxTokens: 2048}
	applyAnthropicReasoning(&input,
		&llms.ReasoningConfig{Effort: llms.ReasoningMedium},
		"anthropic.claude-v2", 2048)

	assert.Nil(t, input.Thinking, "budget thinking stays gated to the reasoning allowlist")
}

func TestApplyAnthropicReasoning_NilConfig(t *testing.T) {
	t.Parallel()

	input := anthropicTextGenerationInput{MaxTokens: 2048, Temperature: 0.8}
	applyAnthropicReasoning(&input, nil, "anthropic.claude-opus-4-7-v1:0", 2048)

	assert.Nil(t, input.Thinking)
	assert.Nil(t, input.OutputConfig)
	assert.EqualValues(t, 0.8, input.Temperature)
}
