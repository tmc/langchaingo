package openai

import (
	"strings"
	"testing"

	"github.com/vxcontrol/langchaingo/llms"
)

func TestDisablingThinkingWhereTheEffortTokenDoesNotReachTheVendor(t *testing.T) {
	t.Parallel()

	for _, model := range []string{
		"glm-5.1", "glm-5.2", "glm-4.6", "minimax-m3", "kimi-k2.6",
		"deepseek-v4-flash", "deepseek-v4-pro",
	} {
		body := sendForWire(t, model, llms.WithReasoningDisabled())
		if !strings.Contains(body, `"thinking":{"type":"disabled"}`) {
			t.Errorf("%s: body carries no disabled thinking object: %s", model, body)
		}
		if strings.Contains(body, "reasoning_effort") {
			t.Errorf("%s: body still carries reasoning_effort: %s", model, body)
		}
	}
}
