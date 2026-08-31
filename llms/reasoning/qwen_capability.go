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
	m := strings.TrimPrefix(strings.ToLower(model), "dashscope/")
	if strings.Contains(m, "/") {
		return false
	}
	return qwenParameterCountName.MatchString(m)
}

// ErrThinkingRequiresStream reports thinking asked of a stream-only model
// without a stream.
type ErrThinkingRequiresStream struct{ Model string }

func (e *ErrThinkingRequiresStream) Error() string {
	return fmt.Sprintf("reasoning on model %q is available only on a streaming call", e.Model)
}
