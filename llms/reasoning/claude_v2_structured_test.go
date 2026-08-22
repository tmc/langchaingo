package reasoning_test

import (
	"testing"

	"github.com/vxcontrol/langchaingo/llms/reasoning"
)

func TestClaudeV2RejectsStructuredOutputOnEverySpelling(t *testing.T) {
	t.Parallel()

	for _, model := range []string{
		"claude-2", "claude-2.1",
		"claude-v2", "claude-v2:1",
		"anthropic.claude-v2", "anthropic.claude-v2:1",
		"us.anthropic.claude-v2:1",
	} {
		if reasoning.ClaudeSupportsStructuredOutput(model) {
			t.Errorf("%s: structured output must be refused locally, got supported", model)
		}
	}
}
