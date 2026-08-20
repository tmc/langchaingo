package reasoning

import "strings"

// LikelyReasoningModel reports whether a model probably reasons: everything
// IsReasoningModel knows, plus any name from a family whose current generations
// reason and which is not a known pre-reasoning generation.
//
// For UI hints only. Wire paths keep using IsReasoningModel: a wrong guess here
// costs a control in a form, there it would change the request.
func LikelyReasoningModel(model string) bool {
	if IsReasoningModel(model) {
		return true
	}

	m := strings.ToLower(model)
	if idx := strings.LastIndex(m, "/"); idx != -1 {
		m = m[idx+1:]
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
	}
	return false
}

// openAIPreReasoning matches the GPT generations below the GPT-5 line, where
// reasoning starts.
func openAIPreReasoning(m string) bool {
	return strings.HasPrefix(m, "gpt-3") || strings.HasPrefix(m, "gpt-4")
}

// claudePreThinking matches the Claude generations below 3.7, where extended
// thinking starts.
func claudePreThinking(m string) bool {
	return strings.HasPrefix(m, "claude-2") ||
		strings.HasPrefix(m, "claude-v2") ||
		strings.HasPrefix(m, "claude-instant") ||
		strings.HasPrefix(m, "claude-3")
}

// isOSeries matches the OpenAI o-series naming (o1, o3, o4-mini, successors);
// every generation of it reasons.
func isOSeries(m string) bool {
	return len(m) > 1 && m[0] == 'o' && m[1] >= '1' && m[1] <= '9'
}
