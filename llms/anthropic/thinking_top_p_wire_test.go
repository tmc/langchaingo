package anthropic_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vxcontrol/langchaingo/llms"
)

func TestThinkingKeepsTopPAboveTheFloor(t *testing.T) {
	t.Parallel()

	budget := func(extra ...llms.CallOption) []llms.CallOption {
		return append([]llms.CallOption{
			llms.WithReasoning(llms.ReasoningNone, 1024), llms.WithMaxTokens(4096),
		}, extra...)
	}
	adaptive := func(extra ...llms.CallOption) []llms.CallOption {
		return append([]llms.CallOption{
			llms.WithReasoning(llms.ReasoningHigh, 0), llms.WithMaxTokens(4096),
		}, extra...)
	}

	for _, tc := range []struct {
		name     string
		model    string
		opts     []llms.CallOption
		wantTopP any
		wantTemp any
	}{
		{"бюджет: top_p выше порога вытесняет служебную температуру", "claude-sonnet-4-5",
			budget(llms.WithTopP(0.97)), 0.97, nil},
		{"бюджет: ровно порог", "claude-sonnet-4-5",
			budget(llms.WithTopP(0.95)), 0.95, nil},
		{"бюджет: ниже порога снимается, температура остаётся", "claude-sonnet-4-5",
			budget(llms.WithTopP(0.5)), nil, 1.0},
		{"бюджет: вызывающий задал оба — побеждает температура", "claude-sonnet-4-5",
			budget(llms.WithTopP(0.97), llms.WithTemperature(0.3)), nil, 1.0},
		{"бюджет: без top_p служебная температура на месте", "claude-sonnet-4-5",
			budget(), nil, 1.0},
		{"4.6 на бюджетном пути: top_p выше порога доезжает", "claude-opus-4-6",
			adaptive(llms.WithTopP(0.97)), 0.97, nil},
		{"4.6 на бюджетном пути: ниже порога снимается", "claude-opus-4-6",
			adaptive(llms.WithTopP(0.5)), nil, 1.0},
		{"модель без сэмплинга не получает ни того ни другого", "claude-sonnet-5",
			adaptive(llms.WithTopP(0.97)), nil, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			body, _ := captureMessagesRequestModel(t, tc.model, tc.opts...)
			if tc.wantTopP == nil {
				assert.NotContains(t, body, "top_p")
			} else {
				assert.Equal(t, tc.wantTopP, body["top_p"])
			}
			if tc.wantTemp == nil {
				assert.NotContains(t, body, "temperature")
			} else {
				assert.Equal(t, tc.wantTemp, body["temperature"])
			}
		})
	}
}
