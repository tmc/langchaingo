package reasoning

import "strings"

// ClaudeReasoningKind classifies how a Claude model accepts extended thinking.
// It is the single source of truth for adaptive-vs-budget thinking and whether
// sampling params are permitted, so every provider path resolves the wire shape
// the same way instead of re-deriving it from scattered model-string checks.
type ClaudeReasoningKind int

const (
	// ClaudeReasoningUnknown is any model not explicitly classified below — a
	// non-Claude model, a Claude model without extended thinking, or a Claude
	// generation newer than this table. It is handled as literal pass-through
	// (the caller's requested mechanism is sent unchanged), preserving prior
	// behavior so an unclassified model never regresses.
	ClaudeReasoningUnknown ClaudeReasoningKind = iota
	// ClaudeReasoningAdaptiveOnly is the newest generation (Opus 4.7/4.8, Sonnet
	// 5, Fable 5, Mythos 5): it accepts thinking.type=adaptive + output_config
	// only, and rejects budget_tokens and temperature/top_p with a 400.
	ClaudeReasoningAdaptiveOnly
	// ClaudeReasoningAdaptiveAndBudget is the transitional generation (Opus 4.6,
	// Sonnet 4.6): it accepts both adaptive and budget thinking and permits
	// sampling params.
	ClaudeReasoningAdaptiveAndBudget
	// ClaudeReasoningBudgetOnly is the extended-thinking generation before
	// adaptive existed (Opus 4.5/4.1/4.0, Sonnet 4.5/4.0/3.7, Haiku 4.5): it
	// accepts thinking.type=enabled + budget_tokens and rejects adaptive.
	ClaudeReasoningBudgetOnly
)

// adaptiveOnlyClaude, dualClaude, and budgetOnlyClaude are the explicit model
// sets. Substrings match both the first-party IDs (claude-opus-4-7) and the
// Bedrock IDs (us.anthropic.claude-opus-4-7). Add a new model to exactly one
// set when it launches; anything absent is treated as ClaudeReasoningUnknown.
var (
	adaptiveOnlyClaude = []string{
		"claude-opus-4-7", "claude-opus-4-8",
		"claude-sonnet-5", "claude-fable-5", "claude-mythos-5",
	}
	dualClaude = []string{
		"claude-opus-4-6", "claude-sonnet-4-6",
	}
	budgetOnlyClaude = []string{
		"claude-opus-4-5", "claude-opus-4-1", "claude-opus-4-0", "claude-opus-4-2025",
		"claude-sonnet-4-5", "claude-sonnet-4-0", "claude-sonnet-4-2025",
		"claude-3-7", "claude-3.7",
		"claude-haiku-4-5", "claude-haiku-4.5",
	}
)

// ClaudeReasoningKindFor classifies a Claude model string. Matching is
// case-insensitive and substring-based so provider/region prefixes and
// -vN:0 suffixes do not affect the result.
func ClaudeReasoningKindFor(model string) ClaudeReasoningKind {
	m := strings.ToLower(model)
	if containsAny(m, adaptiveOnlyClaude) {
		return ClaudeReasoningAdaptiveOnly
	}
	if containsAny(m, dualClaude) {
		return ClaudeReasoningAdaptiveAndBudget
	}
	if containsAny(m, budgetOnlyClaude) {
		return ClaudeReasoningBudgetOnly
	}
	return ClaudeReasoningUnknown
}

// ClaudeSupportsThinking reports whether the model is a known extended-thinking
// Claude generation (any tier except Unknown).
func ClaudeSupportsThinking(model string) bool {
	return ClaudeReasoningKindFor(model) != ClaudeReasoningUnknown
}

// ResolveClaudeAdaptive returns whether to send adaptive thinking (true) or
// budget thinking (false) for a Claude model, given the caller's preference
// (adaptivePreferred is true when the caller used WithAdaptiveReasoning).
//
// The rule is deterministic and honors the caller's preference whenever the
// model supports it, falling back only where the preferred mechanism would be
// rejected:
//   - AdaptiveOnly  → always adaptive (budget would 400).
//   - BudgetOnly    → always budget   (adaptive would 400).
//   - AdaptiveAndBudget / Unknown → the caller's preference, unchanged.
//
// So a currently-accepted call keeps its mechanism; only a currently-rejected
// (400) combination is redirected to the mechanism the model accepts.
func ResolveClaudeAdaptive(model string, adaptivePreferred bool) bool {
	switch ClaudeReasoningKindFor(model) {
	case ClaudeReasoningAdaptiveOnly:
		return true
	case ClaudeReasoningBudgetOnly:
		return false
	default: // AdaptiveAndBudget, Unknown
		return adaptivePreferred
	}
}

// ClaudeRejectsSampling reports whether the model rejects temperature/top_p
// outright (true only for the adaptive-only generation), so sampling params
// must be dropped even when no thinking is requested.
func ClaudeRejectsSampling(model string) bool {
	return ClaudeReasoningKindFor(model) == ClaudeReasoningAdaptiveOnly
}

func containsAny(s string, subs []string) bool {
	for _, sub := range subs {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}
