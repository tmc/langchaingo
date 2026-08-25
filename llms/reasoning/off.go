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
			// Only Anthropic-operated platforms accept thinking:{disabled} here. On
			// Amazon Bedrock the adaptive-only default-on models (e.g. Sonnet 5) keep
			// thinking always on, so a disable is rejected — report it as unsupported.
			if p == ProviderBedrock || p == ProviderOpenAI {
				return OffUnsupported
			}
			return OffDisableClaude
		default:
			return OffOmit // off by default (Opus 4.7/4.8, 4.6, budget-only): omit == off
		}
	}

	switch p {
	case ProviderGoogleAI:
		// Gemini 3.x and Gemini 2.5 Pro cannot be disabled (budget:0 is ignored /
		// rejected); Gemini 2.5 Flash / Flash-Lite, Gemma 4, and unknowns disable
		// via thinkingBudget:0.
		if !GeminiCanDisable(model) {
			return OffUnsupported
		}
		// Known non-thinking families reject thinkingBudget:0, so omit rather than
		// send it; unknown Gemini/Gemma names stay optimistic (attempt budget:0).
		if geminiKnownNonThinking(model) || GeminiThinkingOffByDefault(model) {
			return OffOmit
		}
		return OffZeroBudget
	case ProviderOpenAI:
		if !IsReasoningModel(model) {
			return OffOmit // non-reasoning model does not think
		}
		// A model known to reject disabling (o-series, GPT-5 Pro) is unsupported;
		// unknown models stay optimistic and attempt "none", with a 400 backstop.
		if caps := OpenAIReasoningCapsFor(model); caps.Known && !caps.CanDisable {
			return OffUnsupported
		}
		return OffEffortNone
	default:
		// Bedrock non-Claude and everything else: no clean disable signal — omit.
		return OffOmit
	}
}

func isClaudeModel(model string) bool {
	return strings.Contains(baseModelName(model), "claude")
}

// openAIMandatoryReasoning reports the OpenAI reasoning families that cannot be
// hard-disabled (their effort floor is minimal/low). GPT-5.x and unknowns are
// treated as disablable and left to the API to reject if wrong.
func openAIMandatoryReasoning(model string) bool {
	for _, form := range modelSpellings(model) {
		if strings.HasPrefix(form, "o1") ||
			strings.HasPrefix(form, "o3") ||
			strings.HasPrefix(form, "o4-mini") {
			return true
		}
	}
	return false
}
