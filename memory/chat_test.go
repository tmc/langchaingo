package memory

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/0xDezzy/langchaingo/llms"
)

func TestChatMessageHistory(t *testing.T) {
	ctx := context.Background()
	t.Parallel()

	h := NewChatMessageHistory()
	err := h.AddAIMessage(ctx, "foo")
	require.NoError(t, err)
	err = h.AddUserMessage(ctx, "bar")
	require.NoError(t, err)

	messages, err := h.Messages(ctx)
	require.NoError(t, err)

	assert.Equal(t, []llms.ChatMessage{
		llms.AIChatMessage{Content: "foo"},
		llms.HumanChatMessage{Content: "bar"},
	}, messages)

	h = NewChatMessageHistory(
		WithPreviousMessages([]llms.ChatMessage{
			llms.AIChatMessage{Content: "foo"},
			llms.SystemChatMessage{Content: "bar"},
		}),
	)
	err = h.AddUserMessage(ctx, "zoo")
	require.NoError(t, err)

	messages, err = h.Messages(ctx)
	require.NoError(t, err)

	assert.Equal(t, []llms.ChatMessage{
		llms.AIChatMessage{Content: "foo"},
		llms.SystemChatMessage{Content: "bar"},
		llms.HumanChatMessage{Content: "zoo"},
	}, messages)
}
