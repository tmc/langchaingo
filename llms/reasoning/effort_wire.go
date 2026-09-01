package reasoning

import "strings"

// RejectsPenalties reports whether frequency_penalty and presence_penalty
// must stay off the wire.
func RejectsPenalties(model string) bool {
	for _, form := range modelSpellings(model) {
		if strings.HasPrefix(form, "grok") {
			return true
		}
	}
	return false
}

// UsesLegacyMaxTokens reports whether the output limit must travel as
// max_tokens rather than max_completion_tokens.
func UsesLegacyMaxTokens(model string) bool {
	for _, form := range modelSpellings(model) {
		if strings.HasPrefix(form, "grok") ||
			strings.HasPrefix(form, "qwen") ||
			strings.HasPrefix(form, "qwq") ||
			strings.HasPrefix(form, "qvq") {
			return true
		}
	}
	return false
}

// AcceptsEffortWire reports whether the model's OpenAI-compatible door takes the
// reasoning_effort field. A door that refuses it fails the whole request, so such
// a model also has no way to spell "off": the disable token rides on this field.
func AcceptsEffortWire(model string) bool {
	for _, form := range modelSpellings(model) {
		if strings.HasPrefix(form, "qwen") || strings.HasPrefix(form, "qwq") {
			return false
		}
		if strings.HasPrefix(form, "grok-code-fast") ||
			strings.HasPrefix(form, "grok-build") ||
			strings.HasPrefix(form, "grok-4.20") {
			return false
		}
	}
	return true
}
