package reasoning

import (
	"slices"
	"strings"
)

// openAIEffortRank orders OpenAI reasoning efforts so a requested effort can be
// clamped to a model's ceiling. It does not include "none", which is the disable
// token handled by ResolveOff, not an effort level.
var openAIEffortRank = map[string]int{
	"minimal": 1,
	"low":     2,
	"medium":  3,
	"high":    4,
	"xhigh":   5,
	"max":     6,
}

// OpenAIReasoningCaps is a best-effort static projection of which reasoning
// efforts an OpenAI model accepts, so callers avoid sending a value the API would
// reject with a 400. It asserts only documented, non-default constraints; every
// other model returns Known=false and is treated optimistically (send as
// requested, let the API be the arbiter), preserving prior pass-through behavior.
type OpenAIReasoningCaps struct {
	// Known reports whether the model was explicitly classified.
	Known bool
	// CanDisable reports whether the model accepts reasoning_effort "none".
	CanDisable bool
	// Efforts are the accepted effort levels in ascending order, excluding "none";
	// nil when unknown.
	Efforts []string
}

// ClampEffort moves a requested effort to one the model accepts: above the
// ceiling drops to the ceiling, below the floor rises to the floor, and anything
// the set does not list steps down to the nearest level it does. Unknown models
// and efforts outside the scale are returned unchanged.
func (c OpenAIReasoningCaps) ClampEffort(effort string) string {
	want, ranked := openAIEffortRank[effort]
	if !c.Known || !ranked || len(c.Efforts) == 0 {
		return effort
	}
	if slices.Contains(c.Efforts, effort) {
		return effort
	}

	accepted := c.Efforts[0]
	for _, level := range c.Efforts {
		if openAIEffortRank[level] <= want {
			accepted = level
		}
	}
	return accepted
}

// OpenAIReasoningCapsFor classifies an OpenAI reasoning model by the effort set
// its generation accepts on /chat/completions. A model outside the listed
// generations (including newer ones) returns Known=false so the caller stays
// optimistic and the API arbitrates.
func OpenAIReasoningCapsFor(model string) OpenAIReasoningCaps {
	m := strings.ToLower(model)
	if idx := strings.LastIndex(m, "/"); idx != -1 {
		m = m[idx+1:] // strip a proxy prefix such as "openai/"
	}
	if strings.Contains(m, "-chat-latest") {
		return OpenAIReasoningCaps{Known: false}
	}
	switch {
	case openAIProVariant(m):
		return OpenAIReasoningCaps{Known: true, CanDisable: false, Efforts: []string{"high"}}
	case openAIMandatoryReasoning(m):
		return OpenAIReasoningCaps{Known: true, CanDisable: false, Efforts: []string{"low", "medium", "high", "xhigh"}}
	case openAIGPT5Base(m):
		return OpenAIReasoningCaps{Known: true, CanDisable: false, Efforts: []string{"minimal", "low", "medium", "high"}}
	case strings.HasPrefix(m, "gpt-5.1"):
		return OpenAIReasoningCaps{Known: true, CanDisable: true, Efforts: []string{"low", "medium", "high"}}
	case openAIXHighCeiling(m):
		return OpenAIReasoningCaps{Known: true, CanDisable: true, Efforts: []string{"low", "medium", "high", "xhigh"}}
	default:
		return OpenAIReasoningCaps{Known: false}
	}
}

func openAIProVariant(m string) bool {
	return strings.HasPrefix(m, "gpt-5") && strings.Contains(m, "-pro")
}

func openAIGPT5Base(m string) bool {
	return m == "gpt-5" || strings.HasPrefix(m, "gpt-5-20") ||
		strings.HasPrefix(m, "gpt-5-mini") || strings.HasPrefix(m, "gpt-5-nano")
}

func openAIXHighCeiling(m string) bool {
	for _, prefix := range []string{"gpt-5.2", "gpt-5.4", "gpt-5.5", "gpt-5.6"} {
		if strings.HasPrefix(m, prefix) {
			return true
		}
	}
	return false
}

// OpenAIDisableEffort turns thinking off on the wire. It is not llms.ReasoningNone,
// which is the empty string and instead omits the field.
const OpenAIDisableEffort = "none"

// OpenAIAcceptsCustomTemperature reports whether a reasoning model takes a
// temperature other than the default, so the caller's value is not overwritten
// on a model that honors it.
func OpenAIAcceptsCustomTemperature(model string) bool {
	m := strings.ToLower(model)
	if idx := strings.LastIndex(m, "/"); idx != -1 {
		m = m[idx+1:]
	}
	for _, prefix := range []string{"gpt-5.1", "gpt-5.2", "gpt-5.4"} {
		if strings.HasPrefix(m, prefix) {
			return true
		}
	}
	return false
}
