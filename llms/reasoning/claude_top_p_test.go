package reasoning

import "testing"

func TestClaudeKeepsTopPWhileThinking(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		model string
		topP  float64
		want  bool
	}{
		{"claude-sonnet-4-5", 0.97, true},
		{"claude-sonnet-4-5", 0.95, true},
		{"claude-sonnet-4-5", 0.94, false},
		{"claude-sonnet-4-5", 0.5, false},
		{"claude-opus-4-6", 0.97, true},
		{"anthropic/claude-sonnet-4-5", 0.97, true},
		{"claude-sonnet-5", 0.97, false},
		{"claude-opus-4-7", 0.97, false},
		{"claude-fable-5-1", 0.97, false},
		{"gpt-5.4", 0.97, false},
		{"o3", 0.97, false},
		{"gemini-3.5-flash", 0.97, false},
	} {
		if got := ClaudeKeepsTopPWhileThinking(tc.model, tc.topP); got != tc.want {
			t.Errorf("ClaudeKeepsTopPWhileThinking(%q, %v) = %v, want %v",
				tc.model, tc.topP, got, tc.want)
		}
	}
}
