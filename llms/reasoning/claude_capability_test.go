package reasoning

import (
	"slices"
	"testing"
)

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
		{"claude-opus-5", ClaudeReasoningAdaptiveOnly},
		{"us.anthropic.claude-opus-5", ClaudeReasoningAdaptiveOnly},
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

func TestClaudeEffortsFor(t *testing.T) {
	t.Parallel()
	cases := []struct {
		model string
		want  []string
	}{
		{"claude-opus-5", []string{"low", "medium", "high", "xhigh", "max"}},
		{"us.anthropic.claude-opus-4-7", []string{"low", "medium", "high", "xhigh", "max"}},
		{"claude-opus-4-6", []string{"low", "medium", "high", "max"}},
		{"claude-sonnet-4-6", []string{"low", "medium", "high", "max"}},
		{"claude-opus-4-5-20251101", []string{"low", "medium", "high"}},
		{"gpt-5.5", nil},
	}
	for _, tc := range cases {
		if got := ClaudeEffortsFor(tc.model); !slices.Equal(got, tc.want) {
			t.Errorf("ClaudeEffortsFor(%q) = %v, want %v", tc.model, got, tc.want)
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
		{"claude-opus-5", false, true, "budget on Opus 5 upgrades"},
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
		{"claude-opus-5", true},
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

// TestClaudeThinkingDefaultAndAlwaysOn locks Opus 5's distinguishing trait: it
// joins Sonnet 5 in defaulting to thinking on (unlike Opus 4.8, which defaults
// off), yet — unlike Fable 5 / Mythos 5 — it still accepts an explicit disable.
func TestClaudeThinkingDefaultAndAlwaysOn(t *testing.T) {
	t.Parallel()
	cases := []struct {
		model      string
		defaultOn  bool
		alwaysOn   bool
		reasonNote string
	}{
		{"claude-opus-5", true, false, "defaults on, still disablable"},
		{"claude-sonnet-5", true, false, "defaults on, still disablable"},
		{"claude-opus-4-8", false, false, "defaults off"},
		{"claude-fable-5", true, true, "always on, cannot disable"},
		{"claude-mythos-5", true, true, "always on, cannot disable"},
	}
	for _, tc := range cases {
		if got := ClaudeThinkingDefaultsOn(tc.model); got != tc.defaultOn {
			t.Errorf("ClaudeThinkingDefaultsOn(%q) = %v, want %v (%s)", tc.model, got, tc.defaultOn, tc.reasonNote)
		}
		if got := ClaudeThinkingAlwaysOn(tc.model); got != tc.alwaysOn {
			t.Errorf("ClaudeThinkingAlwaysOn(%q) = %v, want %v (%s)", tc.model, got, tc.alwaysOn, tc.reasonNote)
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
		{"claude-opus-5", true},
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

func TestClaudePredatesAdaptive(t *testing.T) {
	t.Parallel()
	cases := []struct {
		model string
		want  bool
	}{
		{"claude-2.1", true},
		{"anthropic.claude-v2", true},
		{"anthropic.claude-v2:1", true},
		{"anthropic.claude-instant-v1", true},
		{"claude-3-5-haiku", true},
		{"claude-3-7-sonnet-20250219", true},
		{"claude-opus-4-1-20250805", true},
		{"claude-opus-4-20250514", true},
		{"us.anthropic.claude-sonnet-4-20250514-v1:0", true},
		{"claude-opus-4-5", false},
		{"claude-sonnet-4-5", false},
		{"claude-haiku-4-5", false},
		{"claude-opus-4-6", false},
		{"claude-sonnet-4-6", false},
		{"claude-opus-4-8", false},
		{"claude-sonnet-5", false},
		{"us.anthropic.claude-fable-5", false},
	}
	for _, tc := range cases {
		if got := ClaudePredatesAdaptive(tc.model); got != tc.want {
			t.Errorf("ClaudePredatesAdaptive(%q) = %v, want %v", tc.model, got, tc.want)
		}
	}
}
