package reasoning_test

import (
	"testing"

	"github.com/vxcontrol/langchaingo/llms/reasoning"
)

func TestResolveMechanism(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name         string
		model        string
		adaptive     bool
		claudeFamily bool
		reasons      bool
		want         reasoning.Mechanism
	}{
		{
			"a known thinking Claude follows the model, not the flag",
			"anthropic.claude-sonnet-4-5-20250929-v1:0", false, true, true, reasoning.MechanismBudget,
		},
		{
			"an adaptive-only Claude ignores a budget request",
			"anthropic.claude-opus-4-7-20260115-v1:0", false, true, true, reasoning.MechanismAdaptive,
		},
		{
			"an unclassified Claude honors an explicit adaptive request",
			"anthropic.claude-sonnet-9-20991231-v1:0", true, true, true, reasoning.MechanismAdaptive,
		},
		{
			"a pre-adaptive Claude never gets adaptive",
			"anthropic.claude-3-5-haiku-20241022-v1:0", true, true, true, reasoning.MechanismBudget,
		},
		{
			"a non-Claude family with no mechanism of its own gets nothing",
			"openai.gpt-oss-120b-1:0", true, false, true, reasoning.MechanismNone,
		},
		{
			"a model the door does not consider reasoning gets nothing",
			"amazon.nova-lite-v1:0", true, false, false, reasoning.MechanismNone,
		},
	} {
		if got := reasoning.ResolveMechanism(tc.model, tc.adaptive, tc.claudeFamily, tc.reasons); got != tc.want {
			t.Errorf("%s: ResolveMechanism(%q) = %v, want %v", tc.name, tc.model, got, tc.want)
		}
	}
}
