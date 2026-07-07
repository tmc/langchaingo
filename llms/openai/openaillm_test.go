package openai

import (
	"testing"

	"github.com/tmc/langchaingo/llms"
)

func TestGetModelCapabilities(t *testing.T) {
	tests := []struct {
		model        string
		wantThinking bool
		wantSystem   bool
	}{
		{"o1-mini", true, false},
		{"o3", true, false},
		{"o4-mini", true, true},
		{"o3-2025-04-16", true, true},
		{"gpt-5", true, true},
		{"gpt-5-mini", true, true},
		{"gpt-4o", false, true},
		{"gpt-4", false, true},
		{"gpt-3.5-turbo", false, true},
	}
	for _, tt := range tests {
		caps := getModelCapabilities(tt.model)
		if caps.SupportsThinking != tt.wantThinking {
			t.Errorf("getModelCapabilities(%q).SupportsThinking = %v, want %v", tt.model, caps.SupportsThinking, tt.wantThinking)
		}
		if caps.SupportsSystem != tt.wantSystem {
			t.Errorf("getModelCapabilities(%q).SupportsSystem = %v, want %v", tt.model, caps.SupportsSystem, tt.wantSystem)
		}
	}
}

func TestClampReasoningEffort(t *testing.T) {
	tests := []struct {
		effort llms.ThinkingEffort
		want   string
	}{
		{"low", "low"},
		{"medium", "medium"},
		{"high", "high"},
		// OpenAI accepts xhigh, so it must not be downgraded.
		{"xhigh", "xhigh"},
		// max is Anthropic-only; it clamps to OpenAI's highest.
		{"max", "xhigh"},
		{"", ""},
	}
	for _, tt := range tests {
		if got := clampReasoningEffort(tt.effort); got != tt.want {
			t.Errorf("clampReasoningEffort(%q) = %q, want %q", tt.effort, got, tt.want)
		}
	}
}
