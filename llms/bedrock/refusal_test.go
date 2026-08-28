package bedrock_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/llms/bedrock"
)

func TestLegacyRefusalReachesTheCallerAsATypedError(t *testing.T) {
	t.Parallel()

	const resp = `{"id":"x","type":"message","role":"assistant","model":"m",` +
		`"content":[{"type":"text","text":"I can't help with that."}],` +
		`"stop_reason":"refusal",` +
		`"stop_details":{"category":"cyber","explanation":"content boundary"},` +
		`"usage":{"input_tokens":11,"output_tokens":7,"cache_creation_input_tokens":3,"cache_read_input_tokens":5}}`

	llm, _ := legacyLLMCapturing(t, resp, bedrock.WithModel(claudeBudgetModel))
	got, err := llm.GenerateContent(context.Background(),
		[]llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "hi")})

	var refusal *llms.ErrModelRefusal
	require.ErrorAs(t, err, &refusal)
	assert.Equal(t, "I can't help with that.", refusal.Message)
	assert.Equal(t, "cyber", refusal.Category, "stop_details must survive the door")
	assert.Equal(t, "content boundary", refusal.Explanation)
	assert.Equal(t, 11, refusal.InputTokens)
	assert.Equal(t, 7, refusal.OutputTokens)
	assert.Equal(t, 3, refusal.CacheCreationInputTokens)
	assert.Equal(t, 5, refusal.CacheReadInputTokens)

	require.NotNil(t, got, "the response travels with the error so usage survives")
}
