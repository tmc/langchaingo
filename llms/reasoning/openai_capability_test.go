package reasoning

import (
	"slices"
	"testing"
)

func TestOpenAIReasoningCapsFor(t *testing.T) {
	t.Parallel()
	cases := []struct {
		model      string
		known      bool
		canDisable bool
		efforts    []string
	}{
		{"gpt-5-pro", true, false, []string{"high"}},
		{"gpt-5.2-pro", true, false, []string{"medium", "high", "xhigh"}},
		{"gpt-5.4-pro", true, false, []string{"medium", "high", "xhigh"}},
		{"gpt-5.5-pro", true, false, []string{"medium", "high", "xhigh"}},
		{"o3-mini", true, false, []string{"low", "medium", "high", "xhigh"}},
		{"o1-preview", true, false, []string{"low", "medium", "high", "xhigh"}},
		{"o4-mini", true, false, []string{"low", "medium", "high", "xhigh"}},
		{"gpt-5", true, false, []string{"minimal", "low", "medium", "high"}},
		{"gpt-5-2025-08-07", true, false, []string{"minimal", "low", "medium", "high"}},
		{"gpt-5-mini", true, false, []string{"minimal", "low", "medium", "high"}},
		{"gpt-5-nano", true, false, []string{"minimal", "low", "medium", "high"}},
		{"gpt-5.1", true, true, []string{"low", "medium", "high"}},
		{"gpt-5.2", true, true, []string{"low", "medium", "high", "xhigh"}},
		{"gpt-5.4", true, true, []string{"low", "medium", "high", "xhigh"}},
		{"gpt-5.4-mini", true, true, []string{"low", "medium", "high", "xhigh"}},
		{"openai/gpt-5.4-mini", true, true, []string{"low", "medium", "high", "xhigh"}}, // proxy prefix stripped
		{"gpt-5.5", true, true, []string{"low", "medium", "high", "xhigh"}},
		{"gpt-5.6-terra", true, true, []string{"low", "medium", "high", "xhigh"}},
		{"gpt-5.7", false, false, nil}, // unclassified -> optimistic
		{"gpt-4.1", false, false, nil},
	}
	for _, tc := range cases {
		caps := OpenAIReasoningCapsFor(tc.model)
		if caps.Known != tc.known {
			t.Errorf("%q Known = %v, want %v", tc.model, caps.Known, tc.known)
		}
		if caps.Known && caps.CanDisable != tc.canDisable {
			t.Errorf("%q CanDisable = %v, want %v", tc.model, caps.CanDisable, tc.canDisable)
		}
		if !slices.Equal(caps.Efforts, tc.efforts) {
			t.Errorf("%q Efforts = %v, want %v", tc.model, caps.Efforts, tc.efforts)
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
		{"o3", "xhigh", "xhigh"},
		{"o4-mini", "max", "xhigh"},
		{"gpt-5", "xhigh", "high"},
		{"gpt-5-mini", "minimal", "minimal"},
		{"gpt-5.1", "xhigh", "high"},
		{"gpt-5.4-mini", "max", "xhigh"},
		{"gpt-5.6-terra", "xhigh", "xhigh"},
		{"gpt-5.4-mini", "medium", "medium"},
		// Unknown model: unchanged (optimistic).
		{"gpt-5.7", "max", "max"},
		// Empty effort untouched.
		{"gpt-5-pro", "", ""},
	}
	for _, tc := range cases {
		if got := OpenAIReasoningCapsFor(tc.model).ClampEffort(tc.effort); got != tc.want {
			t.Errorf("ClampEffort(%q, %q) = %q, want %q", tc.model, tc.effort, got, tc.want)
		}
	}
}

func TestResolveOffOpenAIGeneration(t *testing.T) {
	t.Parallel()
	cases := []struct {
		model string
		want  OffWire
	}{
		{"gpt-5", OffUnsupported},
		{"gpt-5-mini", OffUnsupported},
		{"gpt-5-nano", OffUnsupported},
		{"o3", OffUnsupported},
		{"gpt-5.1", OffEffortNone},
		{"gpt-5.4-mini", OffEffortNone},
		{"gpt-5.6-terra", OffEffortNone},
	}
	for _, tc := range cases {
		if got := ResolveOff(tc.model, ProviderOpenAI); got != tc.want {
			t.Errorf("ResolveOff(%q) = %v, want %v", tc.model, got, tc.want)
		}
	}
}

func TestClampEffortMovesToAnAcceptedLevel(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		model  string
		effort string
		want   string
	}{
		{"gpt-5.2", "minimal", "low"},
		{"gpt-5.2", "max", "xhigh"},
		{"gpt-5.2", "medium", "medium"},
		{"gpt-5", "minimal", "minimal"},
		{"gpt-5", "xhigh", "high"},
		{"gpt-5-pro", "low", "high"},
		{"gpt-5-pro", "max", "high"},
		{"gpt-5.1", "xhigh", "high"},
		{"claude-opus-4-6", "xhigh", "xhigh"},
		{"gpt-5.2", "", ""},
		{"gpt-5.2", "turbo", "turbo"},
	} {
		if got := OpenAIReasoningCapsFor(tc.model).ClampEffort(tc.effort); got != tc.want {
			t.Errorf("%s.ClampEffort(%q) = %q, want %q", tc.model, tc.effort, got, tc.want)
		}
	}
}

func TestAGenerationDoesNotAnswerForALaterOne(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		model string
		known bool
		why   string
	}{
		{"gpt-5.1", true, "the generation itself"},
		{"gpt-5.10", false, "a later generation whose name extends gpt-5.1"},
		{"gpt-5.2", true, "the generation itself"},
		{"gpt-5.20", false, "a later generation whose name extends gpt-5.2"},
		{"gpt-5", true, "the generation itself"},
		{"gpt-51", false, "a name that merely starts with gpt-5"},
		{"o1", true, "the generation itself"},
		{"o10", false, "a later generation whose name extends o1"},
		{"o3", true, "the generation itself"},
		{"o30", false, "a later generation whose name extends o3"},
		{"gpt-5-2025-08-07", true, "a dated snapshot: the digits after gpt-5-20 are a date, not a generation"},
		{"gpt-5-mini-2025-08-07", true, "a dated snapshot of a named variant"},
		{"o4-mini-2025-04-16", true, "a dated snapshot of the o-series"},
	} {
		t.Run(tc.model, func(t *testing.T) {
			t.Parallel()
			if got := OpenAIReasoningCapsFor(tc.model).Known; got != tc.known {
				t.Errorf("OpenAIReasoningCapsFor(%q).Known = %v, want %v — %s",
					tc.model, got, tc.known, tc.why)
			}
		})
	}
}

func TestTheDocumentedMaxEffortIsNotOnTheWireForGPT56(t *testing.T) {
	t.Parallel()

	for _, model := range []string{"gpt-5.6", "gpt-5.6-sol", "gpt-5.6-terra"} {
		caps := OpenAIReasoningCapsFor(model)
		if got := caps.ClampEffort("max"); got != "xhigh" {
			t.Errorf("ClampEffort(%q, max) = %q, want xhigh: the vendor enumerates its own set as "+
				"none, low, medium, high, xhigh and refuses max with 400", model, got)
		}
		if slices.Contains(caps.Efforts, "max") {
			t.Errorf("%s advertises max, which the vendor refuses", model)
		}
	}
}
