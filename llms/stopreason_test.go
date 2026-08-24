package llms

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsTruncatedCoversEveryVendorSpelling(t *testing.T) {
	t.Parallel()

	truncated := map[string]string{
		"openai":           "length",
		"mistral":          "length",
		"ollama":           "length",
		"anthropic":        "max_tokens",
		"bedrock converse": "max_tokens",
		"bedrock nova":     "max_tokens",
		"bedrock amazon":   "LENGTH",
		"bedrock ai21":     "length",
		"bedrock cohere":   "MAX_TOKENS",
		"bedrock meta":     "length",
		"bedrock deepseek": "length",
		"googleai":         "MAX_TOKENS",
		"vertex":           "MAX_TOKENS",
		"mistral context":  "model_length",
	}
	for vendor, reason := range truncated {
		assert.True(t, IsTruncated(reason), "%s spells truncation %q", vendor, reason)
	}

	finished := []string{
		"stop", "end_turn", "tool_use", "tool_calls", "function_call", "stop_sequence",
		"COMPLETE", "FINISH", "endoftext", "content_filtered", "CONTENT_FILTERED",
		"ERROR", "guardrail_intervened", "", "max_tokens_exceeded", "no_length",
	}
	for _, reason := range finished {
		assert.False(t, IsTruncated(reason), "%q does not mean truncation", reason)
	}
}

func TestCheckTruncationStaysSilentUnlessRequested(t *testing.T) {
	t.Parallel()

	cut := &ContentResponse{Choices: []*ContentChoice{
		{Content: "half an ans", StopReason: "max_tokens", Truncated: true},
	}}
	whole := &ContentResponse{Choices: []*ContentChoice{
		{Content: "a whole answer", StopReason: "end_turn"},
	}}

	require.NoError(t, CheckTruncation(cut, CallOptions{}))
	require.NoError(t, CheckTruncation(whole, CallOptions{FailOnTruncation: true}))
	require.NoError(t, CheckTruncation(nil, CallOptions{FailOnTruncation: true}))

	err := CheckTruncation(cut, CallOptions{FailOnTruncation: true})
	require.Error(t, err)
	assert.True(t, IsTruncatedError(err))
	assert.True(t, errors.Is(err, ErrTruncated))
	assert.Contains(t, err.Error(), "max_tokens")
}

func TestCheckTruncationReadsTheStopReasonNotTheFlag(t *testing.T) {
	t.Parallel()

	unflagged := &ContentResponse{Choices: []*ContentChoice{
		{Content: "half an ans", StopReason: "length"},
	}}

	err := CheckTruncation(unflagged, CallOptions{FailOnTruncation: true})
	require.Error(t, err)
	assert.True(t, IsTruncatedError(err))
}

func TestCheckTruncationReportsALaterChoice(t *testing.T) {
	t.Parallel()

	resp := &ContentResponse{Choices: []*ContentChoice{
		nil,
		{Content: "fine", StopReason: "stop"},
		{Content: "cut", StopReason: "length", Truncated: true},
	}}

	err := CheckTruncation(resp, CallOptions{FailOnTruncation: true})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "choice 2")
}
