package llms

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

// HasAssistantPrefill reports whether the conversation ends with an assistant
// turn, which some models reject.
func HasAssistantPrefill(messages []MessageContent) bool {
	if len(messages) == 0 {
		return false
	}
	return messages[len(messages)-1].Role == ChatMessageTypeAI
}
