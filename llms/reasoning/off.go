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
	// OffDisableDashScope → DashScope enable_thinking:false.
	OffDisableDashScope
	// OffDisableThinkingObject → thinking:{type:"disabled"} on an OpenAI-shaped door.
	OffDisableThinkingObject
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
// best-effort disable wire rather than a silent omit, so a caller's explicit
// "off" is honored wherever the provider can honor it.
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
		if mandatoryThinking(model) {
			return OffUnsupported
		}
		if offByOmission(model) {
			return OffOmit
		}
		if QwenThinkingRequiresStream(model) || QwenThinkingOffByFlag(model) {
			return OffDisableDashScope
		}
		if disablesByThinkingObject(model) {
			return OffDisableThinkingObject
		}
		// The disable token rides on the effort field, so a door that refuses that
		// field cannot express "off" at all.
		if caps := OpenAIReasoningCapsFor(model); caps.Known && !caps.CanDisable {
			return OffUnsupported
		}
		if !AcceptsEffortWire(model) {
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
		if hasGeneration(form, "o1") ||
			hasGeneration(form, "o3") ||
			hasGeneration(form, "o4-mini") {
			return true
		}
	}
	return false
}

func offByOmission(model string) bool {
	if QwenThinkingEnabledByFlag(model) {
		return true
	}
	for _, form := range modelSpellings(model) {
		if strings.HasPrefix(form, "magistral") {
			return true
		}
	}
	return false
}

func disablesByThinkingObject(model string) bool {
	for _, form := range modelSpellings(model) {
		if hasGeneration(form, "minimax-m3") {
			return true
		}
		if hasGeneration(form, "glm-5.2") {
			continue
		}
		for _, generation := range []string{"glm-4.5", "glm-4.6", "glm-4.7", "glm-5", "glm-5.1"} {
			if hasGeneration(form, generation) {
				return true
			}
		}
	}
	return false
}

func mandatoryThinking(model string) bool {
	for _, form := range modelSpellings(model) {
		if strings.HasPrefix(form, "glm-5.3") ||
			form == "grok-build-latest" ||
			strings.HasPrefix(form, "grok-4.5") ||
			strings.HasPrefix(form, "grok-4.6") ||
			strings.HasPrefix(form, "minimax-m2") ||
			form == "glm-latest" ||
			form == "glm-flash-latest" ||
			strings.HasPrefix(form, "kimi-k2-thinking") ||
			strings.HasPrefix(form, "kimi-k2.7-code") ||
			strings.HasPrefix(form, "gpt-oss") ||
			strings.HasPrefix(form, "aion-") ||
			strings.HasPrefix(form, "step-3") ||
			strings.HasPrefix(form, "reka-flash-3") ||
			strings.HasPrefix(form, "fugu-ultra") ||
			strings.HasPrefix(form, "nex-n2") ||
			strings.HasPrefix(form, "lfm-2.5") ||
			strings.HasPrefix(form, "deepseek-r1") ||
			strings.HasPrefix(form, "deepseek-reasoner") ||
			strings.Contains(form, "trinity-large-thinking") {
			return true
		}
	}
	return false
}
