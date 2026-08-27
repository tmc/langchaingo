package anthropic_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/vxcontrol/langchaingo/llms"
)

func TestBudgetNeverFallsBelowTheVendorFloor(t *testing.T) {
	t.Parallel()

	// 1535 is the last max-tokens whose two-thirds cut lands below the floor.
	for _, tc := range []struct {
		maxTokens int
		effort    llms.ReasoningEffort
	}{
		{512, llms.ReasoningLow},
		{512, llms.ReasoningHigh},
		{1024, llms.ReasoningLow},
		{1535, llms.ReasoningLow},
		{1535, llms.ReasoningHigh},
		{1536, llms.ReasoningLow},
		{4096, llms.ReasoningHigh},
	} {
		p, _ := captureMessagesRequestModel(t, "claude-sonnet-4-5",
			llms.WithMaxTokens(tc.maxTokens), llms.WithReasoning(tc.effort, 0))

		thinking, ok := p["thinking"].(map[string]any)
		require.True(t, ok, "maxTokens=%d effort=%s: no thinking block", tc.maxTokens, tc.effort)
		budget, _ := thinking["budget_tokens"].(float64)
		maxTokens, _ := p["max_tokens"].(float64)

		assert.GreaterOrEqual(t, int(budget), 1024,
			"maxTokens=%d effort=%s: budget below the vendor floor", tc.maxTokens, tc.effort)
		assert.Less(t, budget, maxTokens,
			"maxTokens=%d effort=%s: budget must leave room for the answer", tc.maxTokens, tc.effort)
	}
}
