package memory

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/schema"
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

	assert.Equal(t, []schema.ChatMessage{
		schema.AIChatMessage{Content: "foo"},
		schema.HumanChatMessage{Content: "bar"},
	}, messages)

	h = NewChatMessageHistory(
		WithPreviousMessages([]schema.ChatMessage{
			schema.AIChatMessage{Content: "foo"},
			schema.SystemChatMessage{Content: "bar"},
		}),
	)
	err = h.AddUserMessage(context.Background(), "zoo")
	require.NoError(t, err)

	messages, err = h.Messages(context.Background())
	require.NoError(t, err)

	assert.Equal(t, []schema.ChatMessage{
		schema.AIChatMessage{Content: "foo"},
		schema.SystemChatMessage{Content: "bar"},
		schema.HumanChatMessage{Content: "zoo"},
	}, messages)
}
