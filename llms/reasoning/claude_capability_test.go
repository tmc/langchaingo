package reasoning

import "testing"

func TestClaudeReasoningKindFor(t *testing.T) {
	t.Parallel()
	cases := []struct {
		model string
		want  ClaudeReasoningKind
	}{
		// Adaptive-only (T1), first-party and Bedrock forms.
		{"claude-opus-4-7", ClaudeReasoningAdaptiveOnly},
		{"claude-opus-4-8", ClaudeReasoningAdaptiveOnly},
		{"us.anthropic.claude-opus-4-8", ClaudeReasoningAdaptiveOnly},
		{"claude-sonnet-5", ClaudeReasoningAdaptiveOnly},
		{"us.anthropic.claude-sonnet-5", ClaudeReasoningAdaptiveOnly},
		{"claude-fable-5", ClaudeReasoningAdaptiveOnly},
		{"claude-mythos-5", ClaudeReasoningAdaptiveOnly},
		// Dual (T2).
		{"claude-opus-4-6", ClaudeReasoningAdaptiveAndBudget},
		{"us.anthropic.claude-sonnet-4-6", ClaudeReasoningAdaptiveAndBudget},
		// Budget-only (T3).
		{"claude-opus-4-5-20251101", ClaudeReasoningBudgetOnly},
		{"claude-sonnet-4-5-20250929", ClaudeReasoningBudgetOnly},
		{"us.anthropic.claude-haiku-4-5-20251001-v1:0", ClaudeReasoningBudgetOnly},
		// Unknown: no-thinking Claude, retired/non-classified, non-Claude, empty.
		{"claude-3-5-haiku-20241022", ClaudeReasoningUnknown},
		{"claude-2.1", ClaudeReasoningUnknown},
		{"gpt-5.5", ClaudeReasoningUnknown},
		{"", ClaudeReasoningUnknown},
	}
	for _, tc := range cases {
		if got := ClaudeReasoningKindFor(tc.model); got != tc.want {
			t.Errorf("ClaudeReasoningKindFor(%q) = %d, want %d", tc.model, got, tc.want)
		}
	}
}

func TestResolveClaudeAdaptive(t *testing.T) {
	t.Parallel()
	// (model, adaptivePreferred) -> adaptive sent. The invariant: a
	// currently-accepted combination keeps its mechanism; only rejected combos flip.
	cases := []struct {
		model     string
		preferAda bool
		want      bool
		note      string
	}{
		// Adaptive-only: budget preference upgraded to adaptive (M1/M2/H1b).
		{"claude-sonnet-5", false, true, "budget on T1 upgrades"},
		{"claude-sonnet-5", true, true, "adaptive on T1 stays"},
		// Budget-only: adaptive preference downgraded to budget (M3).
		{"claude-haiku-4-5", true, false, "adaptive on T3 downgrades"},
		{"claude-haiku-4-5", false, false, "budget on T3 stays"},
		// Dual: caller preference honored both ways (unchanged).
		{"claude-opus-4-6", true, true, "adaptive on T2 honored"},
		{"claude-opus-4-6", false, false, "budget on T2 honored"},
		// Unknown: literal pass-through (unchanged).
		{"claude-3-5-haiku", true, true, "adaptive on unknown literal"},
		{"claude-3-5-haiku", false, false, "budget on unknown literal"},
	}
	for _, tc := range cases {
		if got := ResolveClaudeAdaptive(tc.model, tc.preferAda); got != tc.want {
			t.Errorf("ResolveClaudeAdaptive(%q, prefer=%v) = %v, want %v (%s)",
				tc.model, tc.preferAda, got, tc.want, tc.note)
		}
	}
}

func TestClaudeSupportsStructuredOutput(t *testing.T) {
	t.Parallel()
	cases := []struct {
		model string
		want  bool
	}{
		{"claude-sonnet-5", true},
		{"claude-opus-4-5", true},
		{"claude-haiku-4-5", true},
		{"us.anthropic.claude-sonnet-4-5-v1:0", true},
		{"some-future-model", true}, // unknown passes through
		{"claude-3-5-sonnet-latest", false},
		{"claude-3-7-sonnet", false},
		{"claude-2.1", false},
		{"claude-instant-1", false},
	}
	for _, tc := range cases {
		if got := ClaudeSupportsStructuredOutput(tc.model); got != tc.want {
			t.Errorf("ClaudeSupportsStructuredOutput(%q) = %v, want %v", tc.model, got, tc.want)
		}
	}
}

func TestClaudeRejectsSampling(t *testing.T) {
	t.Parallel()
	cases := []struct {
		model string
		want  bool
	}{
		{"claude-sonnet-5", true},
		{"us.anthropic.claude-opus-4-8", true},
		{"claude-opus-4-6", false},
		{"claude-sonnet-4-5", false},
		{"claude-3-5-haiku", false},
		{"gpt-5.5", false},
	}
	for _, tc := range cases {
		if got := ClaudeRejectsSampling(tc.model); got != tc.want {
			t.Errorf("ClaudeRejectsSampling(%q) = %v, want %v", tc.model, got, tc.want)
		}
	}
}
