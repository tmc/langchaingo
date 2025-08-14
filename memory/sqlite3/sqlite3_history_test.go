package sqlite3_test

import (
	"context"
	"testing"

	"github.com/0xDezzy/langchaingo/llms"
	"github.com/0xDezzy/langchaingo/memory/sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSqliteChatMessageHistory(t *testing.T) {
	ctx := context.Background()
	t.Parallel()
	h := sqlite3.NewSqliteChatMessageHistory(sqlite3.WithContext(ctx))

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

	h = sqlite3.NewSqliteChatMessageHistory(
		sqlite3.WithContext(ctx),
		sqlite3.WithOverwrite(),
	)

	err = h.SetMessages(ctx,
		[]llms.ChatMessage{
			llms.AIChatMessage{Content: "foo"},
			llms.SystemChatMessage{Content: "bar"},
		})
	require.NoError(t, err)

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
