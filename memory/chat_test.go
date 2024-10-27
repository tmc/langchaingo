package memory

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/starmvp/langchaingo/llms"
)

func TestChatMessageHistory(t *testing.T) {
	t.Parallel()

	h := NewChatMessageHistory()
	err := h.AddAIMessage(context.Background(), "foo")
	require.NoError(t, err)
	err = h.AddUserMessage(context.Background(), "bar")
	require.NoError(t, err)

	messages, err := h.Messages(context.Background())
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
	err = h.AddUserMessage(context.Background(), "zoo")
	require.NoError(t, err)

	messages, err = h.Messages(context.Background())
	require.NoError(t, err)

	assert.Equal(t, []llms.ChatMessage{
		llms.AIChatMessage{Content: "foo"},
		llms.SystemChatMessage{Content: "bar"},
		llms.HumanChatMessage{Content: "zoo"},
	}, messages)
}
