package openai

import (
	"strings"
	"testing"

	"github.com/vxcontrol/langchaingo/llms"
)

func TestClaudeThinkingKeepsTopPAboveTheFloor(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name   string
		model  string
		opts   []llms.CallOption
		want   string
		absent bool
	}{
		{"Claude держит top_p выше порога", "claude-sonnet-4-5",
			[]llms.CallOption{llms.WithTopP(0.97)}, `"top_p":0.97`, false},
		{"Claude держит ровно порог", "claude-sonnet-4-5",
			[]llms.CallOption{llms.WithTopP(0.95)}, `"top_p":0.95`, false},
		{"Claude снимает ниже порога", "claude-sonnet-4-5",
			[]llms.CallOption{llms.WithTopP(0.5)}, `"top_p"`, true},
		{"вызывающий задал оба — top_p уходит", "claude-sonnet-4-5",
			[]llms.CallOption{llms.WithTopP(0.97), llms.WithTemperature(0.3)}, `"top_p"`, true},
		{"модель без сэмплинга не получает top_p", "claude-sonnet-5",
			[]llms.CallOption{llms.WithTopP(0.97)}, `"top_p"`, true},
		{"reasoning-линейка OpenAI по-прежнему теряет top_p", "gpt-5.4",
			[]llms.CallOption{llms.WithTopP(0.97)}, `"top_p"`, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			opts := append([]llms.CallOption{
				llms.WithMaxTokens(8000), llms.WithReasoning(llms.ReasoningHigh, 0),
			}, tc.opts...)
			body := bodyForCall(t, tc.model, opts...)
			if got := strings.Contains(body, tc.want); got == tc.absent {
				verb := "не содержит"
				if tc.absent {
					verb = "содержит"
				}
				t.Errorf("тело запроса %s %s\nтело: %s", verb, tc.want, body)
			}
		})
	}
}
