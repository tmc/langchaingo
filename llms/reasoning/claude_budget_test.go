package reasoning_test

import (
	"testing"

	"github.com/vxcontrol/langchaingo/llms/reasoning"
)

func TestClaudeClampBudget(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name   string
		model  string
		budget int
		want   int
	}{
		{"a budget model is raised to the floor", "claude-sonnet-4-5", 682, 1024},
		{"one below the floor is raised", "claude-sonnet-4-5", 1023, 1024},
		{"the floor itself passes", "claude-sonnet-4-5", 1024, 1024},
		{"a larger budget passes", "claude-sonnet-4-5", 4096, 4096},
		{"a dual model is raised too", "claude-opus-4-6", 500, 1024},
		{"bedrock ids resolve the same", "us.anthropic.claude-sonnet-4-5-20250929-v1:0", 341, 1024},
		{"thinking off passes through", "claude-sonnet-4-5", 0, 0},
		{"an adaptive-only model never sends a budget", "claude-opus-4-8", 341, 341},
		{"a non-Claude model is left alone", "gpt-oss-120b", 341, 341},
	} {
		if got := reasoning.ClaudeClampBudget(tc.model, tc.budget); got != tc.want {
			t.Errorf("%s: ClaudeClampBudget(%q, %d) = %d, want %d", tc.name, tc.model, tc.budget, got, tc.want)
		}
	}
}

func TestClaudeMaxTokensForBudget(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name      string
		budget    int
		maxTokens int
		want      int
	}{
		{"a tight ceiling is raised to twice the budget", 1024, 512, 2048},
		{"a ceiling equal to the budget still leaves no room", 1024, 1024, 2048},
		{"a roomy ceiling is kept", 1024, 8192, 8192},
		{"no budget leaves the ceiling alone", 0, 512, 512},
	} {
		if got := reasoning.ClaudeMaxTokensForBudget(tc.budget, tc.maxTokens); got != tc.want {
			t.Errorf("%s: ClaudeMaxTokensForBudget(%d, %d) = %d, want %d",
				tc.name, tc.budget, tc.maxTokens, got, tc.want)
		}
	}
}
