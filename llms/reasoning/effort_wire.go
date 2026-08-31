package reasoning

import "strings"

// AcceptsEffortWire reports whether the model's OpenAI-compatible door takes the
// reasoning_effort field. A door that refuses it fails the whole request, so such
// a model also has no way to spell "off": the disable token rides on this field.
func AcceptsEffortWire(model string) bool {
	for _, form := range modelSpellings(model) {
		if strings.HasPrefix(form, "qwen") || strings.HasPrefix(form, "qwq") {
			return false
		}
		if strings.Contains(form, "kimi-k2") {
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
