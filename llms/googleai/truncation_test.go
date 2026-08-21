package googleai

import (
	"context"
	"testing"

	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/llms/streaming"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGoogleAITruncationIsReported(t *testing.T) {
	t.Parallel()

	llm := newHTTPRRClient(t, WithDefaultModel("gemini-2.5-flash"))
	msgs := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeHuman, "Describe the water cycle in full detail."),
	}

	cut, err := llm.GenerateContent(t.Context(), msgs, llms.WithMaxTokens(24))
	require.NoError(t, err)
	assert.True(t, cut.Choices[0].Truncated)
	assert.Equal(t, "MAX_TOKENS", cut.Choices[0].StopReason)

	partial, err := llm.GenerateContent(t.Context(), msgs, llms.WithMaxTokens(24), llms.WithFailOnTruncation())
	require.Error(t, err)
	assert.True(t, llms.IsTruncatedError(err))
	require.NotNil(t, partial)

	whole, err := llm.GenerateContent(t.Context(), msgs, llms.WithMaxTokens(4096))
	require.NoError(t, err)
	assert.False(t, whole.Choices[0].Truncated)
	assert.Equal(t, "STOP", whole.Choices[0].StopReason)
}

func TestGoogleAITruncationIsReportedWhenStreaming(t *testing.T) {
	t.Parallel()

	llm := newHTTPRRClient(t, WithDefaultModel("gemini-2.5-flash"))
	msgs := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeHuman, "Describe the water cycle in full detail."),
	}
	sink := func(context.Context, streaming.Chunk) error { return nil }

	cut, err := llm.GenerateContent(t.Context(), msgs,
		llms.WithMaxTokens(200), llms.WithStreamingFunc(sink))
	require.NoError(t, err)
	assert.True(t, cut.Choices[0].Truncated)
	assert.Equal(t, "MAX_TOKENS", cut.Choices[0].StopReason)
}
