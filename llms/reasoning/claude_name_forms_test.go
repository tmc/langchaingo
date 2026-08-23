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
		{"claude-sonnet-4@20250514", "claude-sonnet-4-20250514"},
		{"CLAUDE-Opus-4.6", "claude-opus-4-6"},
		{"gpt-4.1", "gpt-4-1"},
	} {
		if got := canonicalClaude(tc.in); got != tc.want {
			t.Errorf("canonicalClaude(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestDetectionSeesPlatformIdsAndDashedVersions(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		model string
		want  bool
	}{
		{"us.anthropic.claude-opus-5-v1:0", true},
		{"us.anthropic.claude-sonnet-4-6-v1:0", true},
		{"eu.anthropic.claude-opus-4-6-v1:0", true},
		{"claude-3-7-sonnet-20250219", true},
		{"claude-3.7-sonnet", true},
		{"claude-haiku-4-5", true},
		{"glm-4.5", true},
		{"deepseek-v3.1-terminus", true},
		{"us.amazon.titan-text-lite-v1", false},
	} {
		if got := IsReasoningModel(tc.model); got != tc.want {
			t.Errorf("IsReasoningModel(%q) = %v, want %v", tc.model, got, tc.want)
		}
	}
}

func TestStripPlatformPrefixLeavesVersionsAlone(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct{ in, want string }{
		{"us.anthropic.claude-opus-5-v1:0", "claude-opus-5-v1:0"},
		{"anthropic.claude-opus-4-6", "claude-opus-4-6"},
		{"glm-4.5", "glm-4.5"},
		{"deepseek-v3.1-terminus", "deepseek-v3.1-terminus"},
		{"claude-opus-4-7", "claude-opus-4-7"},
	} {
		if got := stripPlatformPrefix(tc.in); got != tc.want {
			t.Errorf("stripPlatformPrefix(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestOpenAISuffixVariantsDoNotInheritTheBaseRules(t *testing.T) {
	t.Parallel()

	for _, model := range []string{"gpt-5-pro", "gpt-5.1-pro", "gpt-5.2-pro", "gpt-5.4-pro"} {
		caps := OpenAIReasoningCapsFor(model)
		if !caps.Known || caps.CanDisable || len(caps.Efforts) != 1 || caps.Efforts[0] != "high" {
			t.Errorf("%s: want the Pro caps (high only, not disablable), got %+v", model, caps)
		}
	}

	for _, model := range []string{"gpt-5-chat-latest", "gpt-5.2-chat-latest"} {
		if caps := OpenAIReasoningCapsFor(model); caps.Known {
			t.Errorf("%s: a chat variant must not be classified as a reasoning model, got %+v", model, caps)
		}
		if IsReasoningModel(model) {
			t.Errorf("%s: a chat variant must not be detected as reasoning", model)
		}
	}
}

func TestClaudeFourAliasesAreClassified(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		model string
		want  bool
	}{
		{"claude-opus-4-0", true},
		{"claude-sonnet-4-0", true},
		{"claude-sonnet-4@20250514", true},
		{"claude-opus-4@20250514", true},
		{"claude-opus-4-20250514", true},
		{"claude-sonnet-4-5", false},
		{"claude-sonnet-4-6", false},
		{"claude-opus-4-6", false},
	} {
		if got := ClaudePredatesAdaptive(tc.model); got != tc.want {
			t.Errorf("ClaudePredatesAdaptive(%q) = %v, want %v", tc.model, got, tc.want)
		}
	}
}
