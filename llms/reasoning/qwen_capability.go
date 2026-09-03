package reasoning

import (
	"fmt"
	"regexp"
	"strings"
)

var qwenParameterCountName = regexp.MustCompile(`^qwen3-\d+(\.\d+)?b(-a\d+(\.\d+)?b)?$`)

var qwenDefaultThinkingName = regexp.MustCompile(`^qwen3\.[567]-`)

var qwenNextName = regexp.MustCompile(`^qwen3-next-.*-thinking$`)

var qwenVLOpenWeightName = regexp.MustCompile(`^qwen3-vl-\d+(\.\d+)?b(-a\d+(\.\d+)?b)?$`)

// QwenThinkingOffByFlag reports whether the model thinks until enable_thinking
// false asks it to stop.
func QwenThinkingOffByFlag(model string) bool {
	return qwenDefaultThinkingName.MatchString(dashScopeSpelling(model))
}

// QwenThinkingRequiresStream reports whether DashScope serves the model's
// thinking only on a streaming call.
func QwenThinkingRequiresStream(model string) bool {
	return qwenParameterCountName.MatchString(dashScopeSpelling(model))
}

func dashScopeSpelling(model string) string {
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
	"qwen-plus": true, "qwen-flash": true, "qwen-turbo": true,
	"qwen3-max": true, "qwen3-vl-plus": true, "qwen3-vl-flash": true,
}

// QwenThinkingEnabledByFlag reports whether DashScope leaves the model's
// thinking off until enable_thinking:true asks for it.
func QwenThinkingEnabledByFlag(model string) bool {
	return qwenFlagThinkers[dashScopeSpelling(model)]
}

var dashScopeGuestBudget = []string{"glm-5.1", "glm-5.2", "kimi-k2.7-code", "deepseek-v4"}

// DashScopeTakesThinkingBudget reports whether DashScope caps the model's
// thinking by a token budget.
func DashScopeTakesThinkingBudget(model string) bool {
	m := dashScopeSpelling(model)
	if m == "" {
		return false
	}
	for _, generation := range dashScopeGuestBudget {
		if hasGeneration(m, generation) {
			return true
		}
	}
	return qwenDefaultThinkingName.MatchString(m) ||
		qwenParameterCountName.MatchString(m) ||
		qwenNextName.MatchString(m) ||
		qwenVLOpenWeightName.MatchString(m) ||
		qwenFlagThinkers[m] ||
		hasGeneration(m, "qwen3.8")
}
