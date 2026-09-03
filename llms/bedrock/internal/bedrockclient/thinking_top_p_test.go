package bedrockclient

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vxcontrol/langchaingo/llms"
)

func TestLegacyThinkingKeepsTopPAboveTheFloor(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name  string
		model string
		topP  float64
		temp  float64
		want  float64
	}{
		{"выше порога доезжает", "anthropic.claude-sonnet-4-5-v1:0", 0.97, 0, 0.97},
		{"вызывающий задал оба — top_p уходит", "anthropic.claude-sonnet-4-5-v1:0", 0.97, 0.3, 0},
		{"ровно порог доезжает", "anthropic.claude-sonnet-4-5-v1:0", 0.95, 0, 0.95},
		{"ниже порога снимается", "anthropic.claude-sonnet-4-5-v1:0", 0.5, 0, 0},
		{"модель без сэмплинга не получает", "anthropic.claude-sonnet-5-v1:0", 0.97, 0, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			input := anthropicTextGenerationInput{MaxTokens: 4096, TopP: tc.topP, Temperature: tc.temp}
			require.NoError(t, applyAnthropicReasoning(&input,
				&llms.ReasoningConfig{Tokens: 1024}, tc.model, 4096))
			assert.InDelta(t, tc.want, input.TopP, 1e-9)
		})
	}
}

func TestConverseThinkingKeepsTopPAboveTheFloor(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name  string
		model string
		topP  float64
		temp  *float64
		want  *float32
	}{
		{"выше порога доезжает", "anthropic.claude-sonnet-4-5-v1:0", 0.97, nil, ptr(float32(0.97))},
		{"вызывающий задал оба — top_p уходит", "anthropic.claude-sonnet-4-5-v1:0", 0.97, ptr(0.3), nil},
		{"ниже порога снимается", "anthropic.claude-sonnet-4-5-v1:0", 0.5, nil, nil},
		{"модель без сэмплинга не получает", "anthropic.claude-sonnet-5-v1:0", 0.97, nil, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			client := NewConverseClient(&MockBedrockRuntimeClient{})
			built, err := client.buildConverseInput(&ConverseInput{
				ModelID:         tc.model,
				Messages:        []Message{{Role: llms.ChatMessageTypeHuman, Content: "hi", Type: "text"}},
				MaxTokens:       ptr(4096),
				TopP:            ptr(tc.topP),
				Temperature:     tc.temp,
				ReasoningConfig: &llms.ReasoningConfig{Tokens: 1024},
			})
			require.NoError(t, err)
			if tc.want == nil {
				assert.Nil(t, built.InferenceConfig.TopP)
				return
			}
			require.NotNil(t, built.InferenceConfig.TopP)
			assert.InDelta(t, *tc.want, *built.InferenceConfig.TopP, 1e-6)
		})
	}
}
