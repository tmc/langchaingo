package bedrock_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/llms/bedrock"
	"github.com/vxcontrol/langchaingo/llms/reasoning"
)

const claudeBudgetModel = "anthropic.claude-sonnet-4-5-20250929-v1:0"

// 1535 is the last max-tokens whose two-thirds cut lands below the floor.
var budgetCases = []struct {
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
}

func TestLegacyBudgetNeverFallsBelowTheVendorFloor(t *testing.T) {
	t.Parallel()

	const resp = `{"id":"x","type":"message","role":"assistant","model":"m",` +
		`"content":[{"type":"text","text":"ok"}],"stop_reason":"end_turn",` +
		`"usage":{"input_tokens":1,"output_tokens":1}}`

	for _, tc := range budgetCases {
		llm, sent := legacyLLMCapturing(t, resp, bedrock.WithModel(claudeBudgetModel))

		_, err := llm.GenerateContent(context.Background(),
			[]llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "hi")},
			llms.WithMaxTokens(tc.maxTokens), llms.WithReasoning(tc.effort, 0))
		require.NoError(t, err)

		var got struct {
			MaxTokens int `json:"max_tokens"`
			Thinking  struct {
				BudgetTokens int `json:"budget_tokens"`
			} `json:"thinking"`
		}
		require.NoError(t, json.Unmarshal([]byte(*sent), &got))

		assert.GreaterOrEqual(t, got.Thinking.BudgetTokens, 1024,
			"maxTokens=%d effort=%s: budget below the vendor floor", tc.maxTokens, tc.effort)
		assert.Less(t, got.Thinking.BudgetTokens, got.MaxTokens,
			"maxTokens=%d effort=%s: budget must leave room for the answer", tc.maxTokens, tc.effort)
	}
}

func TestConverseBudgetNeverFallsBelowTheVendorFloor(t *testing.T) {
	t.Parallel()

	const answer = `{"output":{"message":{"role":"assistant","content":[{"text":"ok"}]}},` +
		`"stopReason":"end_turn","usage":{"inputTokens":1,"outputTokens":1,"totalTokens":2}}`

	for _, tc := range budgetCases {
		var body string
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			b, _ := io.ReadAll(r.Body)
			body = string(b)
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, answer)
		}))

		llm := bedrockLLMAgainst(t, srv, bedrock.WithModel(claudeBudgetModel), bedrock.WithConverseAPI())
		_, err := llm.GenerateContent(context.Background(),
			[]llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "hi")},
			llms.WithMaxTokens(tc.maxTokens), llms.WithReasoning(tc.effort, 0))
		srv.Close()
		require.NoError(t, err)

		var got struct {
			InferenceConfig struct {
				MaxTokens int `json:"maxTokens"`
			} `json:"inferenceConfig"`
			Fields struct {
				Thinking struct {
					BudgetTokens int `json:"budget_tokens"`
				} `json:"thinking"`
			} `json:"additionalModelRequestFields"`
		}
		require.NoError(t, json.Unmarshal([]byte(body), &got))

		assert.GreaterOrEqual(t, got.Fields.Thinking.BudgetTokens, 1024,
			"maxTokens=%d effort=%s: budget below the vendor floor", tc.maxTokens, tc.effort)
		assert.Less(t, got.Fields.Thinking.BudgetTokens, got.InferenceConfig.MaxTokens,
			"maxTokens=%d effort=%s: budget must leave room for the answer", tc.maxTokens, tc.effort)
	}
}

func TestLegacyRefusesAnEffortWithNoBudget(t *testing.T) {
	t.Parallel()

	const resp = `{"id":"x","type":"message","role":"assistant","model":"m",` +
		`"content":[{"type":"text","text":"ok"}],"stop_reason":"end_turn",` +
		`"usage":{"input_tokens":1,"output_tokens":1}}`

	llm, sent := legacyLLMCapturing(t, resp, bedrock.WithModel(claudeBudgetModel))

	_, err := llm.GenerateContent(context.Background(),
		[]llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "hi")},
		llms.WithMaxTokens(8000), llms.WithReasoning(llms.ReasoningMinimal, 0))

	var refused *reasoning.ErrEffortHasNoBudget
	require.ErrorAs(t, err, &refused)
	assert.Equal(t, string(llms.ReasoningMinimal), refused.Effort)
	assert.Empty(t, *sent, "the refusal must land before the request leaves")
}

func TestConverseRefusesAnEffortWithNoBudget(t *testing.T) {
	t.Parallel()

	const answer = `{"output":{"message":{"role":"assistant","content":[{"text":"ok"}]}},` +
		`"stopReason":"end_turn","usage":{"inputTokens":1,"outputTokens":1,"totalTokens":2}}`

	var body string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		body = string(b)
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, answer)
	}))
	t.Cleanup(srv.Close)

	llm := bedrockLLMAgainst(t, srv, bedrock.WithModel(claudeBudgetModel), bedrock.WithConverseAPI())
	_, err := llm.GenerateContent(context.Background(),
		[]llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "hi")},
		llms.WithMaxTokens(8000), llms.WithReasoning(llms.ReasoningMinimal, 0))

	var refused *reasoning.ErrEffortHasNoBudget
	require.ErrorAs(t, err, &refused)
	assert.Equal(t, string(llms.ReasoningMinimal), refused.Effort)
	assert.Empty(t, body, "the refusal must land before the request leaves")
}
