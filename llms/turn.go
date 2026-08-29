package llms

import "github.com/vxcontrol/langchaingo/llms/reasoning"

// ToolChoiceKind is what a tool choice asks of the model, apart from the
// spelling the door expects. Doors translate it into their own wire shape.
type ToolChoiceKind int

const (
	ToolChoiceUnset ToolChoiceKind = iota
	ToolChoiceAuto
	ToolChoiceNone
	ToolChoiceAny
	ToolChoiceNamed
)

// ClassifyToolChoice reads every spelling the doors accept and reports what was
// asked, naming the tool when the caller picked one.
func ClassifyToolChoice(choice any) (ToolChoiceKind, string) {
	kindOf := func(t string, name string) (ToolChoiceKind, string) {
		switch t {
		case "auto":
			return ToolChoiceAuto, ""
		case "none":
			return ToolChoiceNone, ""
		case "tool", "function":
			if name != "" {
				return ToolChoiceNamed, name
			}
			return ToolChoiceAny, ""
		case "any", "required":
			return ToolChoiceAny, ""
		}
		return ToolChoiceUnset, ""
	}

	switch c := choice.(type) {
	case string:
		return kindOf(c, "")
	case ToolChoice:
		return kindOf(c.Type, functionName(c.Function))
	case *ToolChoice:
		if c == nil {
			return ToolChoiceUnset, ""
		}
		return kindOf(c.Type, functionName(c.Function))
	case map[string]any:
		t, _ := c["type"].(string)
		name, _ := c["name"].(string)
		if fn, ok := c["function"].(map[string]any); ok {
			name, _ = fn["name"].(string)
		}
		return kindOf(t, name)
	}
	return ToolChoiceUnset, ""
}

// ForcesToolUse reports whether a tool choice demands a tool call rather than
// leaving the decision to the model.
func ForcesToolUse(choice any) bool {
	kind, _ := ClassifyToolChoice(choice)
	return kind == ToolChoiceAny || kind == ToolChoiceNamed
}

// ForcedToolName reports whether the choice demands a tool call, and names the
// tool when the caller picked one. An empty name with forced=true means "any
// tool", which every door spells differently.
func ForcedToolName(choice any) (name string, forced bool) {
	kind, name := ClassifyToolChoice(choice)
	switch kind { //nolint:exhaustive // the kinds that leave the choice to the model are not forcing
	case ToolChoiceNamed:
		return name, true
	case ToolChoiceAny:
		return "", true
	}
	return "", false
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
	return CheckClaudeTurnLimitsOnWire(model, opts, messages, true)
}

// CheckClaudeTurnLimitsOnWire is CheckClaudeTurnLimits for a door that knows
// whether its own request carries a manual thinking budget; a door that sends
// only an effort passes false.
func CheckClaudeTurnLimitsOnWire(
	model string,
	opts CallOptions,
	messages []MessageContent,
	sendsManualThinking bool,
) error {
	budget := reasoning.ClaudeClampBudget(model, opts.Reasoning.GetTokens(opts.GetMaxTokens()))
	budgetOnly := reasoning.ClaudeReasoningKindFor(model) == reasoning.ClaudeReasoningBudgetOnly
	budgetThinking := (sendsManualThinking || budgetOnly) &&
		opts.Reasoning.ResolveMode() == ReasoningOn &&
		reasoning.ClaudeSupportsThinking(model) &&
		!reasoning.ResolveClaudeAdaptive(model, opts.Reasoning.Adaptive) &&
		budget > 0
	if budgetThinking && ForcesToolUse(opts.ToolChoice) {
		return &reasoning.ErrForcedToolUseWithThinking{Model: model}
	}

	if reasoning.ClaudeRejectsAssistantPrefill(model) && HasAssistantPrefill(messages) {
		return &reasoning.ErrAssistantPrefillUnsupported{Model: model}
	}
	return nil
}
