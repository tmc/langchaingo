package reasoning

import "strings"

// This file is the single source of truth for Google (Gemini / Gemma) reasoning
// classification, so the enable path, the disable path, the reasoning-model
// detector, and the UI hint all agree instead of re-deriving it from scattered
// model-string checks. Provider-wire specifics that need the genai types
// (thinking_level mapping, the temperature value) stay in the googleai adapter.

func baseModelName(model string) string {
	m := strings.ToLower(model)
	if idx := strings.LastIndex(m, "/"); idx != -1 {
		m = m[idx+1:]
	}
	return m
}

func hasFamily(model, family string) bool {
	rest := model
	for {
		idx := strings.Index(rest, family)
		if idx == -1 {
			return false
		}
		rest = rest[idx+len(family):]
		if rest == "" || rest[0] < '0' || rest[0] > '9' {
			return true
		}
	}
}

// GeminiSupportsThinking reports whether the model belongs to a Google thinking
// family: Gemini 2.5, Gemini 3.x, or Gemma 4.
func GeminiSupportsThinking(model string) bool {
	m := baseModelName(model)
	if geminiNonChatSurface(m) {
		return false
	}
	return hasFamily(m, "gemini-2.5") ||
		hasFamily(m, "gemini-3") ||
		hasFamily(m, "gemma-4") ||
		geminiUnversionedThinking(m)
}

func geminiNonChatSurface(model string) bool {
	return strings.HasSuffix(model, "-tts") ||
		strings.Contains(model, "-live-translate") ||
		strings.Contains(model, "-image")
}

func geminiUnversionedThinking(model string) bool {
	for _, prefix := range []string{"gemini-flash-latest", "gemini-flash-lite-latest", "gemini-robotics-er"} {
		if strings.HasPrefix(model, prefix) {
			return true
		}
	}
	return false
}

// GeminiUsesThinkingLevel reports whether the model uses the qualitative
// thinking_level control (Gemini 3.x), where thinking_budget is deprecated,
// instead of a token budget. Gemini 3 also recommends running at temperature 1.0.
func GeminiUsesThinkingLevel(model string) bool {
	return hasFamily(baseModelName(model), "gemini-3")
}

// geminiKnownNonThinking reports whether the model is a pre-thinking Gemini/Gemma
// generation that never thinks (Gemini 1.x/2.0, Gemma 1–3), so it takes no thinking
// control at all. Unclassified names are NOT matched, staying optimistic so a
// future thinking model is not wrongly treated as non-thinking.
func geminiKnownNonThinking(model string) bool {
	m := baseModelName(model)
	return hasFamily(m, "gemini-1") ||
		hasFamily(m, "gemini-2.0") ||
		hasFamily(m, "gemma-1") ||
		hasFamily(m, "gemma-2") ||
		hasFamily(m, "gemma-3")
}

// GeminiCanDisable reports whether thinking can be turned off via
// thinkingBudget:0. Gemini 2.5 Flash/Flash-Lite and Gemma 4 can; Gemini 2.5 Pro
// and Gemini 3.x cannot. Unclassified Google models are treated as disablable,
// so a model this package has not seen is attempted rather than refused.
func GeminiCanDisable(model string) bool {
	m := baseModelName(model)
	if GeminiThinkingOffByDefault(m) {
		return true
	}
	if hasFamily(m, "gemini-3") {
		return false
	}
	if hasFamily(m, "gemini-2.5") && strings.Contains(m, "pro") {
		return false
	}
	return true
}

// GeminiThinkingOffByDefault reports whether the model leaves thinking off until
// asked, so omitting the thinking config already yields "off".
func GeminiThinkingOffByDefault(model string) bool {
	m := baseModelName(model)
	if !strings.Contains(m, "flash-lite") {
		return false
	}
	return strings.Contains(m, "gemini") || strings.Contains(m, "gemma")
}
