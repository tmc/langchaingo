package openaiclient

import (
	"encoding/json"
	"testing"
)

func TestChatRequestReasoningEffort(t *testing.T) {
	tests := []struct {
		name            string
		reasoningEffort ReasoningEffort
		expectedJSON    string
	}{
		{
			name:            "minimal reasoning effort",
			reasoningEffort: ReasoningEffortMinimal,
			expectedJSON:    `"reasoning_effort":"minimal"`,
		},
		{
			name:            "low reasoning effort",
			reasoningEffort: ReasoningEffortLow,
			expectedJSON:    `"reasoning_effort":"low"`,
		},
		{
			name:            "medium reasoning effort",
			reasoningEffort: ReasoningEffortMedium,
			expectedJSON:    `"reasoning_effort":"medium"`,
		},
		{
			name:            "high reasoning effort",
			reasoningEffort: ReasoningEffortHigh,
			expectedJSON:    `"reasoning_effort":"high"`,
		},
		{
			name:            "empty reasoning effort omitted",
			reasoningEffort: "",
			expectedJSON:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := ChatRequest{
				Model:           "gpt-4",
				Messages:        []*ChatMessage{{Role: "user", Content: "test"}},
				Temperature:     0.7,
				ReasoningEffort: tt.reasoningEffort,
			}

			jsonBytes, err := json.Marshal(req)
			if err != nil {
				t.Fatalf("Failed to marshal ChatRequest: %v", err)
			}

			jsonStr := string(jsonBytes)
			if tt.expectedJSON == "" {
				// Should not contain reasoning_effort field
				if !jsonContains(jsonStr, "reasoning_effort") {
					// This is expected - empty reasoning effort should be omitted
				} else {
					t.Errorf("Expected reasoning_effort to be omitted from JSON, but it was present: %s", jsonStr)
				}
			} else {
				// Should contain the expected reasoning_effort value
				if !jsonContains(jsonStr, tt.expectedJSON) {
					t.Errorf("Expected JSON to contain %s, got: %s", tt.expectedJSON, jsonStr)
				}
			}
		})
	}
}

func jsonContains(json, substring string) bool {
	return len(substring) > 0 && len(json) > 0 &&
		(json[0] == '{' || json[0] == '[') &&
		len(json) > len(substring) &&
		indexOf(json, substring) >= 0
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func TestReasoningEffortConstants(t *testing.T) {
	tests := []struct {
		constant ReasoningEffort
		expected string
	}{
		{ReasoningEffortMinimal, "minimal"},
		{ReasoningEffortLow, "low"},
		{ReasoningEffortMedium, "medium"},
		{ReasoningEffortHigh, "high"},
	}

	for _, tt := range tests {
		if string(tt.constant) != tt.expected {
			t.Errorf("Expected constant %v to equal %s, got %s", tt.constant, tt.expected, string(tt.constant))
		}
	}
}
