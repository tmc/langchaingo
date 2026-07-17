package reasoning

import "strings"

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

// ClampEffort lowers a requested effort to what the model accepts: an effort
// above the model's ceiling drops to the ceiling, and a model that accepts a
// single effort (e.g. GPT-5 Pro accepts only "high") pins to it. Unknown models
// and the empty effort are returned unchanged.
func (c OpenAIReasoningCaps) ClampEffort(effort string) string {
	if !c.Known || effort == "" || len(c.Efforts) == 0 {
		return effort
	}
	if len(c.Efforts) == 1 {
		return c.Efforts[0]
	}
	ceiling := c.Efforts[len(c.Efforts)-1]
	if openAIEffortRank[effort] > openAIEffortRank[ceiling] {
		return ceiling
	}
	return effort
}

// OpenAIReasoningCapsFor classifies an OpenAI reasoning model. Only models with
// documented constraints tighter than the general GPT-5 surface are listed;
// anything else (including newer models) returns Known=false so the caller stays
// optimistic.
func OpenAIReasoningCapsFor(model string) OpenAIReasoningCaps {
	m := strings.ToLower(model)
	if idx := strings.LastIndex(m, "/"); idx != -1 {
		m = m[idx+1:] // strip a proxy prefix such as "openai/"
	}
	switch {
	case strings.HasPrefix(m, "gpt-5-pro"):
		// GPT-5 Pro accepts only high and cannot be disabled.
		return OpenAIReasoningCaps{Known: true, CanDisable: false, Efforts: []string{"high"}}
	case openAIMandatoryReasoning(m):
		// o-series floor is minimal/low with no hard off; expose low..high.
		return OpenAIReasoningCaps{Known: true, CanDisable: false, Efforts: []string{"low", "medium", "high"}}
	case strings.HasPrefix(m, "gpt-5.4-mini"):
		// GPT-5.4 mini accepts none..xhigh but not max.
		return OpenAIReasoningCaps{Known: true, CanDisable: true, Efforts: []string{"low", "medium", "high", "xhigh"}}
	default:
		return OpenAIReasoningCaps{Known: false}
	}
}
