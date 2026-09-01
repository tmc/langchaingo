package reasoning

import (
	"fmt"
	"regexp"
	"strings"
)

var qwenParameterCountName = regexp.MustCompile(`^qwen3-\d+(\.\d+)?b(-a\d+(\.\d+)?b)?$`)

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

var qwenFlagThinkers = map[string]bool{"qwen-plus": true, "qwen-flash": true, "qwen3-max": true}

// QwenThinkingEnabledByFlag reports whether DashScope leaves the model's
// thinking off until enable_thinking:true asks for it.
func QwenThinkingEnabledByFlag(model string) bool {
	return qwenFlagThinkers[qwenVendorSpelling(model)]
}
