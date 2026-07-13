package reasoning

import (
	"fmt"
	"strings"
)

// Provider identifies the calling provider so ResolveOff can pick the right
// disable wire for a model. It is passed by the adapter, which always knows it.
type Provider int

const (
	ProviderUnknown Provider = iota
	ProviderAnthropic
	ProviderBedrock
	ProviderOpenAI
	ProviderGoogleAI
)

// OffWire is how a provider expresses "thinking off" for a given model. It is the
// single decision point shared by every adapter, keyed off the same model tables
// the enable path uses, so Off and On can never classify a model differently.
type OffWire int

const (
	// OffOmit sends no reasoning field: the model does not think by default (or the
	// provider needs no explicit signal), so omitting already yields "off".
	OffOmit OffWire = iota
	// OffDisableClaude → Anthropic thinking:{type:"disabled"}.
	OffDisableClaude
	// OffZeroBudget → Google thinkingBudget:0.
	OffZeroBudget
	// OffEffortNone → OpenAI reasoning_effort:"none".
	OffEffortNone
	// OffUnsupported: a known mandatory-thinking model that cannot be disabled
	// (adaptive-only Claude, OpenAI o-series). The adapter returns a typed error.
	OffUnsupported
)

// ErrReasoningOffUnsupported is returned when reasoning is explicitly disabled
// (WithReasoningDisabled) on a model whose thinking cannot be turned off.
type ErrReasoningOffUnsupported struct{ Model string }

func (e *ErrReasoningOffUnsupported) Error() string {
	return fmt.Sprintf("reasoning cannot be disabled for model %q", e.Model)
}

// ResolveOff decides how to disable thinking for a model on a provider, from the
// same capability tables the enable path reads. Unknown models get the provider's
// best-effort disable wire (optimistic: attempt it and let an API error be the
// backstop) rather than a silent omit, so a caller's explicit "off" is honored
// wherever the provider can honor it.
func ResolveOff(model string, p Provider) OffWire {
	if isClaudeModel(model) {
		switch {
		case ClaudeThinkingAlwaysOn(model):
			return OffUnsupported // Fable 5 / Mythos 5: thinking cannot be disabled
		case ClaudeThinkingDefaultsOn(model):
			return OffDisableClaude // on by default (Sonnet 5): needs explicit thinking:{disabled}
		default:
			return OffOmit // off by default (Opus 4.7/4.8, 4.6, budget-only): omit == off
		}
	}

	switch p {
	case ProviderGoogleAI:
		// Gemini 2.5 disables via thinkingBudget:0; Pro / 3.x that reject it surface a 400.
		return OffZeroBudget
	case ProviderOpenAI:
		if !IsReasoningModel(model) {
			return OffOmit // non-reasoning model does not think
		}
		if openAIMandatoryReasoning(model) {
			return OffUnsupported // o-series floor is minimal/low, no hard off
		}
		return OffEffortNone // GPT-5.x accept "none"; unknowns attempt it, 400 backstops
	default:
		// Bedrock non-Claude and everything else: no clean disable signal — omit.
		return OffOmit
	}
}

func isClaudeModel(model string) bool {
	return strings.Contains(strings.ToLower(model), "claude")
}

// openAIMandatoryReasoning reports the OpenAI reasoning families that cannot be
// hard-disabled (their effort floor is minimal/low). GPT-5.x and unknowns are
// treated as disablable and left to the API to reject if wrong.
func openAIMandatoryReasoning(model string) bool {
	m := strings.ToLower(model)
	if idx := strings.LastIndex(m, "/"); idx != -1 {
		m = m[idx+1:]
	}
	return strings.HasPrefix(m, "o1") ||
		strings.HasPrefix(m, "o3") ||
		strings.HasPrefix(m, "o4-mini")
}
