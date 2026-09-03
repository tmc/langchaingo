package reasoning

import (
	"fmt"
	"regexp"
	"strings"
)

var qwenParameterCountName = regexp.MustCompile(`^qwen3-\d+(\.\d+)?b(-a\d+(\.\d+)?b)?$`)

var qwenDefaultThinkingName = regexp.MustCompile(`^qwen3\.[567]-`)

// QwenThinkingOffByFlag reports whether the model thinks until enable_thinking
// false asks it to stop.
func QwenThinkingOffByFlag(model string) bool {
	return qwenDefaultThinkingName.MatchString(qwenVendorSpelling(model))
}

// QwenThinkingRequiresStream reports whether DashScope serves the model's
// thinking only on a streaming call.
func QwenThinkingRequiresStream(model string) bool {
	return qwenParameterCountName.MatchString(qwenVendorSpelling(model))
}

func qwenVendorSpelling(model string) string {
	m := strings.TrimPrefix(strings.ToLower(model), "dashscope/")
	if strings.Contains(m, "/") {
		return ""
	}
	return m
}

// ErrThinkingRequiresStream reports thinking asked of a stream-only model
// without a stream.
type ErrThinkingRequiresStream struct{ Model string }

func (e *ErrThinkingRequiresStream) Error() string {
	return fmt.Sprintf("reasoning on model %q is available only on a streaming call", e.Model)
}

var qwenFlagThinkers = map[string]bool{
	"qwen-plus": true, "qwen-flash": true, "qwen3-max": true,
	"qwen3-vl-plus": true, "qwen3-vl-flash": true,
}

// QwenThinkingEnabledByFlag reports whether DashScope leaves the model's
// thinking off until enable_thinking:true asks for it.
func QwenThinkingEnabledByFlag(model string) bool {
	return qwenFlagThinkers[qwenVendorSpelling(model)]
}

// QwenTakesThinkingBudget reports whether DashScope caps the model's thinking by
// a token budget.
func QwenTakesThinkingBudget(model string) bool {
	m := qwenVendorSpelling(model)
	if m == "" {
		return false
	}
	return qwenDefaultThinkingName.MatchString(m) ||
		qwenParameterCountName.MatchString(m) ||
		qwenFlagThinkers[m] ||
		hasGeneration(m, "qwen3.8")
}
