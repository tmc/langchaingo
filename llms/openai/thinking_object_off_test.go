package openai

import (
	"strings"
	"testing"

	"github.com/vxcontrol/langchaingo/llms"
)

func TestDisablingThinkingWhereTheEffortTokenDoesNotReachTheVendor(t *testing.T) {
	t.Parallel()

	for _, model := range []string{"glm-5.1", "glm-4.6", "minimax-m3"} {
		body := sendForWire(t, model, llms.WithReasoningDisabled())
		if !strings.Contains(body, `"thinking":{"type":"disabled"}`) {
			t.Errorf("%s: body carries no disabled thinking object: %s", model, body)
		}
		if strings.Contains(body, "reasoning_effort") {
			t.Errorf("%s: body still carries reasoning_effort, which the gateway cuts "+
				"for MiniMax and the vendor ignores below GLM-5.2: %s", model, body)
		}
	}

	effortSide := sendForWire(t, "glm-5.2", llms.WithReasoningDisabled())
	if strings.Contains(effortSide, `"thinking"`) {
		t.Errorf("glm-5.2 takes its disable on the effort field, not the object: %s", effortSide)
	}
	if !strings.Contains(effortSide, `"reasoning_effort":"none"`) {
		t.Errorf("glm-5.2 lost its disable token: %s", effortSide)
	}
}
