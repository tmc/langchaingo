package reasoning

import "strings"

// LikelyReasoningModel reports whether a model probably reasons: everything
// IsReasoningModel knows, plus any name from a family whose current generations
// reason and which is not a known pre-reasoning generation.
//
// For UI hints only. Wire paths keep using IsReasoningModel.
func LikelyReasoningModel(model string) bool {
	if IsReasoningModel(model) {
		return true
	}

	m := strings.ToLower(model)
	if idx := strings.LastIndex(m, "/"); idx != -1 {
		m = m[idx+1:]
	}
	m = stripPlatformPrefix(m)

	if namesNonTextModality(m) {
		return false
	}

	switch {
	case strings.HasPrefix(m, "gpt-"):
		return !openAIPreReasoning(m)
	case isOSeries(m):
		return true
	case strings.HasPrefix(m, "claude-"):
		return !claudePreThinking(m)
	case strings.HasPrefix(m, "gemini-"), strings.HasPrefix(m, "gemma-"):
		return !geminiKnownNonThinking(m)
	case strings.HasPrefix(m, "qwen"):
		return !qwenPreReasoning(m)
	case strings.HasPrefix(m, "qvq"):
		return true
	}
	return false
}

func namesNonTextModality(m string) bool {
	return strings.Contains(m, "embedding") || strings.Contains(m, "image")
}

// openAIPreReasoning matches the GPT generations below the GPT-5 line, where
// reasoning starts.
func openAIPreReasoning(m string) bool {
	return strings.HasPrefix(m, "gpt-3") || strings.HasPrefix(m, "gpt-4")
}

// claudePreThinking matches the Claude generations below 3.7, where extended
// thinking starts.
func claudePreThinking(m string) bool {
	c := canonicalClaude(m)
	if strings.Contains(c, "claude-3-7") {
		return false
	}
	return strings.HasPrefix(c, "claude-2") ||
		strings.HasPrefix(c, "claude-v2") ||
		strings.HasPrefix(c, "claude-instant") ||
		strings.HasPrefix(c, "claude-3")
}

func qwenPreReasoning(m string) bool {
	return strings.HasPrefix(m, "qwen1") ||
		strings.HasPrefix(m, "qwen2") ||
		strings.HasPrefix(m, "qwen-1") ||
		strings.HasPrefix(m, "qwen-2")
}

// isOSeries matches the OpenAI o-series naming (o1, o3, o4-mini, successors);
// every generation of it reasons.
func isOSeries(m string) bool {
	return len(m) > 1 && m[0] == 'o' && m[1] >= '1' && m[1] <= '9'
}
