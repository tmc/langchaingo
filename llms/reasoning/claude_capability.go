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
	// ClaudeReasoningAdaptiveOnly is the newest generation (Opus 4.7/4.8/5,
	// Sonnet 5, Fable 5, Mythos 5): it accepts thinking.type=adaptive +
	// output_config only, and rejects budget_tokens and temperature/top_p with a 400.
	ClaudeReasoningAdaptiveOnly
	// ClaudeReasoningAdaptiveAndBudget is the transitional generation (Opus 4.6,
	// Sonnet 4.6): it accepts both adaptive and budget thinking and permits
	// sampling params.
	ClaudeReasoningAdaptiveAndBudget
	// ClaudeReasoningBudgetOnly is the extended-thinking generation before
	// adaptive existed (Opus 4.5, Sonnet 4.5, Haiku 4.5): it accepts
	// thinking.type=enabled + budget_tokens and rejects adaptive.
	ClaudeReasoningBudgetOnly
)

// adaptiveOnlyClaude, dualClaude, and budgetOnlyClaude are the explicit model
// sets. Substrings match both the first-party IDs (claude-opus-4-7) and the
// Bedrock IDs (us.anthropic.claude-opus-4-7). Add a new model to exactly one
// set when it launches; anything absent is treated as ClaudeReasoningUnknown.
var (
	adaptiveOnlyClaude = []string{
		"claude-opus-4-7", "claude-opus-4-8", "claude-opus-5",
		"claude-sonnet-5", "claude-fable-5", "claude-mythos-5",
	}
	dualClaude = []string{
		"claude-opus-4-6", "claude-sonnet-4-6",
	}
	budgetOnlyClaude = []string{
		"claude-opus-4-5",
		"claude-sonnet-4-5",
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

// alwaysOnClaude models think unconditionally and reject an explicit disable.
// defaultOnClaude models think when thinking is omitted (so forcing them off
// needs an explicit disable, not just omission). Opus 5 and Sonnet 5 default to
// thinking on — a breaking change from Opus 4.8, which defaults off — but,
// unlike Fable 5 / Mythos 5, they still accept an explicit disable (subject to
// the API's own effort ceiling: Anthropic rejects thinking.disabled combined
// with effort xhigh/max on Opus 5, a constraint this package does not model
// because this SDK never sends effort alongside a disable request).
var (
	alwaysOnClaude  = []string{"claude-fable-5", "claude-mythos-5"}
	defaultOnClaude = []string{"claude-opus-5", "claude-sonnet-5", "claude-fable-5", "claude-mythos-5"}
)

// ClaudeThinkingAlwaysOn reports whether the model's thinking cannot be disabled
// (Fable 5 / Mythos 5 — an explicit disable returns 400).
func ClaudeThinkingAlwaysOn(model string) bool {
	return containsAny(strings.ToLower(model), alwaysOnClaude)
}

// ClaudeThinkingDefaultsOn reports whether the model thinks when thinking is
// omitted, so disabling it requires an explicit disable rather than omission.
func ClaudeThinkingDefaultsOn(model string) bool {
	return containsAny(strings.ToLower(model), defaultOnClaude)
}

// budgetEffortClaude are budget-thinking models that also accept an effort
// output_config alongside manual thinking (introduced with Opus 4.5). Newer
// generations use adaptive thinking, where effort is always available.
var budgetEffortClaude = []string{"claude-opus-4-5"}

// ClaudeSupportsEffortWithBudget reports whether the model accepts
// output_config.effort together with manual (budget) thinking, so the effort is
// not lost on the budget path for models that honor it.
func ClaudeSupportsEffortWithBudget(model string) bool {
	return containsAny(strings.ToLower(model), budgetEffortClaude)
}

// mutuallyExclusiveSamplingClaude models reject temperature and top_p set
// together (only one may be provided).
var mutuallyExclusiveSamplingClaude = []string{"claude-haiku-4-5", "claude-haiku-4.5"}

// ClaudeMutuallyExclusiveSampling reports whether the model returns a 400 when
// temperature and top_p are set together, so the caller must send at most one.
func ClaudeMutuallyExclusiveSampling(model string) bool {
	return containsAny(strings.ToLower(model), mutuallyExclusiveSamplingClaude)
}

// legacyNoStructuredClaude are Claude generations known to predate structured
// outputs (the output_config.format JSON Schema mode). A structured-output request
// on them is rejected locally rather than sent for a guaranteed 4xx.
var legacyNoStructuredClaude = []string{"claude-2", "claude-instant", "claude-3-", "claude-3."}

// ClaudeSupportsStructuredOutput reports whether the model can be asked for schema
// constrained output. Known-legacy families are rejected; every current model and
// any unrecognized (newer) name passes through so the provider API stays the final
// arbiter and the local table never blocks a future model.
func ClaudeSupportsStructuredOutput(model string) bool {
	return !containsAny(strings.ToLower(model), legacyNoStructuredClaude)
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

// preAdaptiveClaude are known Claude generations released before adaptive thinking
// existed (adaptive arrived with Opus 4.6 / Sonnet 4.6). They are "unknown" to the
// reasoning-kind table only because legacy models are not enumerated there; an
// adaptive request on them must not be forwarded as thinking.type=adaptive, which
// these models reject with a 400.
var preAdaptiveClaude = []string{
	"claude-2", "claude-v2",
	"claude-instant",
	"claude-3", // claude-3, claude-3-5, claude-3-7 all predate adaptive
	"claude-opus-4-1",
	"claude-opus-4-20", "claude-sonnet-4-20",
}

// ClaudePredatesAdaptive reports whether the model is a known pre-adaptive Claude
// generation, so an adaptive request must be gated (not sent verbatim) rather than
// optimistically forwarded the way a genuinely newer, unclassified model is.
func ClaudePredatesAdaptive(model string) bool {
	return containsAny(strings.ToLower(model), preAdaptiveClaude)
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
