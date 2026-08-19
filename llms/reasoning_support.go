package llms

import (
	"strings"
	"sync"

	"github.com/vxcontrol/langchaingo/llms/reasoning"
)

// ReasoningSupport describes which reasoning controls a model accepts, as a HINT
// for building a UI. It is a best-effort static projection, never authoritative:
// an unrecognized model returns Known=false so the UI shows all controls and the
// provider API (an HTTP 400) is the ultimate arbiter. The table asserts only
// KNOWN facts; it never fabricates a capability it cannot back — unknown effort
// tiers and default state are left nil rather than guessed.
type ReasoningSupport struct {
	// Supported reports whether the model reasons at all.
	Supported bool
	// Known reports whether this model is explicitly classified. When false, the
	// remaining fields are best-effort and the UI should offer all controls.
	Known bool
	// CannotDisable is set only for models KNOWN to reject disabling
	// (always-on Claude such as Fable 5, and OpenAI o-series). When false,
	// disabling may still fail at the API for an unclassified model.
	CannotDisable bool
	// RejectsSampling is set for models that reject temperature/top_p while thinking.
	RejectsSampling bool
	// Efforts are the effort tiers worth offering; nil when unknown.
	Efforts []ReasoningEffort
	// DefaultOn reports whether thinking runs when reasoning is unset; nil when unknown.
	DefaultOn *bool
}

var (
	reasoningOverridesMu sync.RWMutex
	reasoningOverrides   = map[string]ReasoningSupport{}
)

// RegisterReasoningSupport registers a UI hint for models this build does not
// classify (a proxy alias, or a model newer than this build). A model whose
// string contains pattern reports info from ReasoningSupportFor.
//
// Scope: this affects ONLY the ReasoningSupportFor hint, not the wire path. The
// enable/disable resolvers (reasoning.ResolveOff, reasoning.ResolveClaudeAdaptive,
// effort clamping) live in the lower-level reasoning package, which cannot read
// this registry, so a registered model still travels the optimistic pass-through
// path on the wire and the provider API remains the final arbiter.
func RegisterReasoningSupport(pattern string, info ReasoningSupport) {
	reasoningOverridesMu.Lock()
	defer reasoningOverridesMu.Unlock()
	reasoningOverrides[strings.ToLower(pattern)] = info
}

func lookupReasoningOverride(model string) (ReasoningSupport, bool) {
	reasoningOverridesMu.RLock()
	defer reasoningOverridesMu.RUnlock()
	m := strings.ToLower(model)
	for pattern, info := range reasoningOverrides {
		if strings.Contains(m, pattern) {
			return info, true
		}
	}
	return ReasoningSupport{}, false
}

// claudeCannotDisable reports whether thinking stays on whatever the caller asks.
// Only the Anthropic adapter emits thinking:{type:"disabled"}.
func claudeCannotDisable(model string, p reasoning.Provider) bool {
	switch reasoning.ResolveOff(model, p) {
	case reasoning.OffUnsupported:
		return true
	case reasoning.OffDisableClaude:
		return p != reasoning.ProviderAnthropic
	default:
		return false
	}
}

func boolPtr(b bool) *bool { return &b }

// ReasoningSupportFor returns the reasoning-control hint for a model on a
// provider. Registered overrides win; then Claude and OpenAI reasoning models are
// classified from the shared tables; everything else returns Known=false
// (optimistic — the UI shows all controls and the API rejects what it cannot do).
func ReasoningSupportFor(model string, p reasoning.Provider) ReasoningSupport {
	if s, ok := lookupReasoningOverride(model); ok {
		return s
	}

	if reasoning.ClaudeSupportsThinking(model) {
		return ReasoningSupport{
			Supported: true,
			Known:     true,
			// Derive from the same resolver the wire uses so the hint tracks
			// provider differences (e.g. Sonnet 5 is disablable on Anthropic but
			// always on, hence not disablable, on Bedrock).
			CannotDisable:   claudeCannotDisable(model, p),
			RejectsSampling: reasoning.ClaudeRejectsSampling(model),
			Efforts:         claudeEfforts(reasoning.ClaudeReasoningKindFor(model)),
			DefaultOn:       boolPtr(reasoning.ClaudeThinkingDefaultsOn(model)),
		}
	}

	if p == reasoning.ProviderOpenAI && reasoning.IsReasoningModel(model) {
		// A classified model (e.g. GPT-5 Pro accepts only high, GPT-5.4 mini adds
		// xhigh) advertises its own effort set; an unclassified one falls back to
		// the conservative low/medium/high triple.
		efforts := []ReasoningEffort{ReasoningLow, ReasoningMedium, ReasoningHigh}
		if caps := reasoning.OpenAIReasoningCapsFor(model); caps.Known {
			efforts = toReasoningEfforts(caps.Efforts)
		}
		return ReasoningSupport{
			Supported:     true,
			Known:         true,
			CannotDisable: reasoning.ResolveOff(model, p) == reasoning.OffUnsupported,
			Efforts:       efforts,
			// DefaultOn is per-model on the GPT-5.x line (some default off) — leave unknown.
		}
	}

	if p == reasoning.ProviderGoogleAI && reasoning.GeminiSupportsThinking(model) {
		return ReasoningSupport{
			// Flash/Flash-Lite/Gemma disable via thinkingBudget:0; Pro and Gemini 3.x
			// cannot fully disable — derive from the same resolver the wire uses.
			CannotDisable: reasoning.ResolveOff(model, p) == reasoning.OffUnsupported,
			Supported:     true,
			Known:         true,
			// These families think by default (Flash-Lite / Gemma may still accept budget:0).
			DefaultOn: boolPtr(true),
		}
	}

	// Unrecognized: optimistic. Supported best-guessed from the shared classifier;
	// the UI shows all controls and the provider teaches via a 400.
	return ReasoningSupport{Supported: reasoning.IsReasoningModel(model), Known: false}
}

// toReasoningEfforts converts the resolver's raw effort strings to the public
// ReasoningEffort type for UI hints.
func toReasoningEfforts(efforts []string) []ReasoningEffort {
	out := make([]ReasoningEffort, len(efforts))
	for i, e := range efforts {
		out[i] = ReasoningEffort(e)
	}
	return out
}

func claudeEfforts(kind reasoning.ClaudeReasoningKind) []ReasoningEffort {
	switch kind {
	case reasoning.ClaudeReasoningAdaptiveOnly:
		return []ReasoningEffort{ReasoningLow, ReasoningMedium, ReasoningHigh, ReasoningXHigh, ReasoningMax}
	case reasoning.ClaudeReasoningAdaptiveAndBudget:
		return []ReasoningEffort{ReasoningLow, ReasoningMedium, ReasoningHigh, ReasoningMax}
	case reasoning.ClaudeReasoningBudgetOnly:
		return []ReasoningEffort{ReasoningLow, ReasoningMedium, ReasoningHigh}
	default:
		return nil
	}
}
