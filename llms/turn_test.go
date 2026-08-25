package llms

import "testing"

func TestForcesToolUseAcceptsEveryFormOfTheChoice(t *testing.T) {
	t.Parallel()

	named := ToolChoice{Type: "tool", Function: &FunctionReference{Name: "calc"}}
	for _, choice := range []any{
		"any", "tool",
		ToolChoice{Type: "any"}, named,
		&ToolChoice{Type: "any"}, &named,
		map[string]any{"type": "any"},
		map[string]any{"type": "tool", "name": "calc"},
	} {
		if !ForcesToolUse(choice) {
			t.Errorf("%#v demands a tool call", choice)
		}
	}

	var absent *ToolChoice
	for _, choice := range []any{
		nil, absent, "auto", "none", "",
		ToolChoice{Type: "auto"}, &ToolChoice{Type: "none"},
		map[string]any{"type": "auto"}, map[string]any{},
		42,
	} {
		if ForcesToolUse(choice) {
			t.Errorf("%#v leaves the decision to the model", choice)
		}
	}
}
