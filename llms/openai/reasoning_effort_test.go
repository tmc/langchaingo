package openai

import (
	"testing"

	"github.com/tmc/langchaingo/llms/openai/internal/openaiclient"
)

func TestReasoningEffortOption(t *testing.T) {
	tests := []struct {
		name            string
		reasoningEffort ReasoningEffort
		expectedAPI     openaiclient.ReasoningEffort
	}{
		{
			name:            "minimal effort",
			reasoningEffort: ReasoningEffortMinimal,
			expectedAPI:     openaiclient.ReasoningEffortMinimal,
		},
		{
			name:            "low effort",
			reasoningEffort: ReasoningEffortLow,
			expectedAPI:     openaiclient.ReasoningEffortLow,
		},
		{
			name:            "medium effort",
			reasoningEffort: ReasoningEffortMedium,
			expectedAPI:     openaiclient.ReasoningEffortMedium,
		},
		{
			name:            "high effort",
			reasoningEffort: ReasoningEffortHigh,
			expectedAPI:     openaiclient.ReasoningEffortHigh,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create options with reasoning effort
			opt := &options{}
			WithReasoningEffort(tt.reasoningEffort)(opt)

			// Verify the option was set correctly
			if opt.reasoningEffort != openaiclient.ReasoningEffort(tt.expectedAPI) {
				t.Errorf("Expected reasoning effort %v, got %v", tt.expectedAPI, opt.reasoningEffort)
			}
		})
	}
}

func TestReasoningEffortConstants(t *testing.T) {
	// Test that our constants match the internal client constants
	if ReasoningEffortMinimal != ReasoningEffort(openaiclient.ReasoningEffortMinimal) {
		t.Errorf("ReasoningEffortMinimal constant mismatch")
	}
	if ReasoningEffortLow != ReasoningEffort(openaiclient.ReasoningEffortLow) {
		t.Errorf("ReasoningEffortLow constant mismatch")
	}
	if ReasoningEffortMedium != ReasoningEffort(openaiclient.ReasoningEffortMedium) {
		t.Errorf("ReasoningEffortMedium constant mismatch")
	}
	if ReasoningEffortHigh != ReasoningEffort(openaiclient.ReasoningEffortHigh) {
		t.Errorf("ReasoningEffortHigh constant mismatch")
	}
}
