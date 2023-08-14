package memory

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tmc/langchaingo/schema"
)

func TestChatMessageHistory(t *testing.T) {
	t.Parallel()

	h := NewChatMessageHistory()
	err := h.AddAIMessage(context.Background(), "foo")
	assert.NoError(t, err)
	err = h.AddUserMessage(context.Background(), "bar")
	assert.NoError(t, err)

	messages, err := h.Messages(context.Background())
	assert.NoError(t, err)

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
	assert.NoError(t, err)

	messages, err = h.Messages(context.Background())
	assert.NoError(t, err)

	assert.Equal(t, []schema.ChatMessage{
		schema.AIChatMessage{Content: "foo"},
		schema.SystemChatMessage{Content: "bar"},
		schema.HumanChatMessage{Content: "zoo"},
	}, messages)
}
