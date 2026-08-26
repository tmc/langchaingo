package llms

import "github.com/vxcontrol/langchaingo/llms/reasoning"

// ForcesToolUse reports whether a tool choice demands a tool call rather than
// leaving the decision to the model.
func ForcesToolUse(choice any) bool {
	forced := func(t string) bool { return t == "any" || t == "tool" }
	switch c := choice.(type) {
	case string:
		return forced(c)
	case ToolChoice:
		return forced(c.Type)
	case *ToolChoice:
		return c != nil && forced(c.Type)
	case map[string]any:
		t, _ := c["type"].(string)
		return forced(t)
	}
	return false
}

// ForcedToolName reports whether the choice demands a tool call, and names the
// tool when the caller picked one. An empty name with forced=true means "any
// tool", which every door spells differently.
func ForcedToolName(choice any) (name string, forced bool) {
	if !ForcesToolUse(choice) {
		return "", false
	}
	switch c := choice.(type) {
	case ToolChoice:
		return functionName(c.Function), true
	case *ToolChoice:
		return functionName(c.Function), true
	case map[string]any:
		if fn, ok := c["function"].(map[string]any); ok {
			n, _ := fn["name"].(string)
			return n, true
		}
		n, _ := c["name"].(string)
		return n, true
	}
	return "", true
}

func functionName(fn *FunctionReference) string {
	if fn == nil {
		return ""
	}
	return fn.Name
}

// HasAssistantPrefill reports whether the conversation ends with an assistant
// turn, which some models reject.
func HasAssistantPrefill(messages []MessageContent) bool {
	if len(messages) == 0 {
		return false
	}
	return messages[len(messages)-1].Role == ChatMessageTypeAI
}

// CheckClaudeTurnLimits refuses the two turns a Claude model rejects on the
// wire: manual (budget) thinking combined with a forced tool choice, and a
// conversation that ends on an assistant turn.
func CheckClaudeTurnLimits(model string, opts CallOptions, messages []MessageContent) error {
	budgetThinking := opts.Reasoning.ResolveMode() == ReasoningOn &&
		reasoning.ClaudeSupportsThinking(model) &&
		!reasoning.ResolveClaudeAdaptive(model, opts.Reasoning.Adaptive)
	if budgetThinking && ForcesToolUse(opts.ToolChoice) {
		return &reasoning.ErrForcedToolUseWithThinking{Model: model}
	}

	if reasoning.ClaudeRejectsAssistantPrefill(model) && HasAssistantPrefill(messages) {
		return &reasoning.ErrAssistantPrefillUnsupported{Model: model}
	}
	return nil
}
