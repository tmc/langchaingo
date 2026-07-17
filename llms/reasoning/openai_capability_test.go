package reasoning

import "testing"

func TestOpenAIReasoningCapsFor(t *testing.T) {
	t.Parallel()
	cases := []struct {
		model      string
		known      bool
		canDisable bool
		ceiling    string // highest accepted effort, "" when unknown
	}{
		{"gpt-5-pro", true, false, "high"},
		{"o3-mini", true, false, "high"},
		{"o1-preview", true, false, "high"},
		{"gpt-5.4-mini", true, true, "xhigh"},
		{"openai/gpt-5.4-mini", true, true, "xhigh"}, // proxy prefix stripped
		{"gpt-5.6", false, false, ""},                // unclassified -> optimistic
		{"gpt-5.5", false, false, ""},
		{"gpt-4.1", false, false, ""},
	}
	for _, tc := range cases {
		caps := OpenAIReasoningCapsFor(tc.model)
		if caps.Known != tc.known {
			t.Errorf("%q Known = %v, want %v", tc.model, caps.Known, tc.known)
		}
		if caps.Known && caps.CanDisable != tc.canDisable {
			t.Errorf("%q CanDisable = %v, want %v", tc.model, caps.CanDisable, tc.canDisable)
		}
		ceiling := ""
		if len(caps.Efforts) > 0 {
			ceiling = caps.Efforts[len(caps.Efforts)-1]
		}
		if ceiling != tc.ceiling {
			t.Errorf("%q ceiling = %q, want %q", tc.model, ceiling, tc.ceiling)
		}
	}
}

func TestOpenAIReasoningCaps_ClampEffort(t *testing.T) {
	t.Parallel()
	cases := []struct {
		model  string
		effort string
		want   string
	}{
		// GPT-5 Pro pins any effort to its single accepted value.
		{"gpt-5-pro", "low", "high"},
		{"gpt-5-pro", "max", "high"},
		{"gpt-5-pro", "high", "high"},
		// GPT-5.4 mini caps at xhigh; max drops to xhigh, lower values pass.
		{"gpt-5.4-mini", "max", "xhigh"},
		{"gpt-5.4-mini", "xhigh", "xhigh"},
		{"gpt-5.4-mini", "medium", "medium"},
		// Unknown model: unchanged (optimistic).
		{"gpt-5.6", "max", "max"},
		{"gpt-5.5", "xhigh", "xhigh"},
		// Empty effort untouched.
		{"gpt-5-pro", "", ""},
	}
	for _, tc := range cases {
		if got := OpenAIReasoningCapsFor(tc.model).ClampEffort(tc.effort); got != tc.want {
			t.Errorf("ClampEffort(%q, %q) = %q, want %q", tc.model, tc.effort, got, tc.want)
		}
	}
}
