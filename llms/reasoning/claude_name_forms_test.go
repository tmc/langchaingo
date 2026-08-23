package reasoning

import "testing"

func TestClaudePredicatesAgreeAcrossNameForms(t *testing.T) {
	t.Parallel()

	for _, group := range [][]string{
		{"claude-opus-4-7", "claude-opus-4.7", "anthropic/claude-opus-4-7", "us.anthropic.claude-opus-4-7-v1:0"},
		{"claude-sonnet-4-5", "claude-sonnet-4.5", "anthropic/claude-sonnet-4.5", "claude-sonnet-4-5@20250514"},
		{"claude-opus-4-6", "claude-opus-4.6", "eu.anthropic.claude-opus-4-6-v1:0"},
		{"claude-haiku-4-5", "claude-haiku-4.5"},
		{"claude-sonnet-5", "claude-sonnet-5.0", "us.anthropic.claude-sonnet-5-v1:0"},
	} {
		canonical := group[0]
		want := struct {
			kind    ClaudeReasoningKind
			reject  bool
			mutex   bool
			prefill bool
		}{
			ClaudeReasoningKindFor(canonical),
			ClaudeRejectsSampling(canonical),
			ClaudeMutuallyExclusiveSampling(canonical),
			ClaudeRejectsAssistantPrefill(canonical),
		}
		if want.kind == ClaudeReasoningUnknown {
			t.Fatalf("%s must be classified for this table to mean anything", canonical)
		}

		for _, form := range group[1:] {
			t.Run(form, func(t *testing.T) {
				t.Parallel()
				if got := ClaudeReasoningKindFor(form); got != want.kind {
					t.Errorf("ClaudeReasoningKindFor = %v, want %v (as %s)", got, want.kind, canonical)
				}
				if got := ClaudeRejectsSampling(form); got != want.reject {
					t.Errorf("ClaudeRejectsSampling = %v, want %v (as %s)", got, want.reject, canonical)
				}
				if got := ClaudeMutuallyExclusiveSampling(form); got != want.mutex {
					t.Errorf("ClaudeMutuallyExclusiveSampling = %v, want %v (as %s)", got, want.mutex, canonical)
				}
				if got := ClaudeRejectsAssistantPrefill(form); got != want.prefill {
					t.Errorf("ClaudeRejectsAssistantPrefill = %v, want %v (as %s)", got, want.prefill, canonical)
				}
			})
		}
	}
}

func TestCanonicalClaudeLeavesUnrelatedDotsAlone(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct{ in, want string }{
		{"claude-opus-4.7", "claude-opus-4-7"},
		{"us.anthropic.claude-opus-4-7-v1:0", "us.anthropic.claude-opus-4-7-v1:0"},
		{"claude-sonnet-4@20250514", "claude-sonnet-4"},
		{"CLAUDE-Opus-4.6", "claude-opus-4-6"},
		{"gpt-4.1", "gpt-4-1"},
	} {
		if got := canonicalClaude(tc.in); got != tc.want {
			t.Errorf("canonicalClaude(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
