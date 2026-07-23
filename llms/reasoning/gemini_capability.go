package reasoning

import "strings"

// This file is the single source of truth for Google (Gemini / Gemma) reasoning
// classification, so the enable path, the disable path, the reasoning-model
// detector, and the UI hint all agree instead of re-deriving it from scattered
// model-string checks. Provider-wire specifics that need the genai types
// (thinking_level mapping, the temperature value) stay in the googleai adapter.

// GeminiSupportsThinking reports whether the model belongs to a Google thinking
// family: Gemini 2.5, Gemini 3.x, or Gemma 4.
func GeminiSupportsThinking(model string) bool {
	m := strings.ToLower(model)
	return strings.Contains(m, "gemini-2.5") ||
		strings.Contains(m, "gemini-3") ||
		strings.Contains(m, "gemma-4")
}

// GeminiUsesThinkingLevel reports whether the model uses the qualitative
// thinking_level control (Gemini 3.x), where thinking_budget is deprecated,
// instead of a token budget. Gemini 3 also recommends running at temperature 1.0.
func GeminiUsesThinkingLevel(model string) bool {
	return strings.Contains(strings.ToLower(model), "gemini-3")
}

// geminiKnownNonThinking reports whether the model is a pre-thinking Gemini/Gemma
// generation that never thinks (Gemini 1.x/2.0, Gemma 1–3), so it takes no thinking
// control at all. Unclassified names are NOT matched, staying optimistic so a
// future thinking model is not wrongly treated as non-thinking.
func geminiKnownNonThinking(model string) bool {
	m := strings.ToLower(model)
	return strings.Contains(m, "gemini-1") ||
		strings.Contains(m, "gemini-2.0") ||
		strings.Contains(m, "gemma-1") ||
		strings.Contains(m, "gemma-2") ||
		strings.Contains(m, "gemma-3")
}

// GeminiCanDisable reports whether thinking can be turned off via
// thinkingBudget:0. Gemini 2.5 Flash/Flash-Lite and Gemma 4 can; Gemini 2.5 Pro
// and Gemini 3.x (budget:0 is ignored) cannot. Unclassified Google models are
// treated as disablable (optimistic: attempt it, let the API be the backstop).
func GeminiCanDisable(model string) bool {
	m := strings.ToLower(model)
	if strings.Contains(m, "gemini-3") {
		return false
	}
	if strings.Contains(m, "gemini-2.5") && strings.Contains(m, "pro") {
		return false
	}
	return true
}
