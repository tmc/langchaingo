package sqlite3_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/memory/sqlite3"
)

func TestSetMessagesEmpty(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	history := sqlite3.NewSqliteChatMessageHistory(sqlite3.WithOverwrite())
	t.Cleanup(func() { require.NoError(t, history.DB.Close()) })
	require.NoError(t, history.AddUserMessage(ctx, "old message"))
	require.NoError(t, history.SetMessages(ctx, nil))
	messages, err := history.Messages(ctx)
	require.NoError(t, err)
	require.Empty(t, messages)
	require.NoError(t, history.AddUserMessage(ctx, "new message"))
	require.NoError(t, history.SetMessages(ctx, []llms.ChatMessage{}))
	messages, err = history.Messages(ctx)
	require.NoError(t, err)
	require.Empty(t, messages)
}

func TestSetMessagesRollsBackOnInsertError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	history := sqlite3.NewSqliteChatMessageHistory(
		sqlite3.WithOverwrite(),
		sqlite3.WithSchema([]byte(`CREATE TABLE langchaingo_messages (
			id INTEGER PRIMARY KEY, session TEXT, type TEXT,
			content TEXT CHECK(content <> 'reject'), created TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);`)),
	)
	t.Cleanup(func() { require.NoError(t, history.DB.Close()) })
	require.NoError(t, history.AddUserMessage(ctx, "keep this message"))
	require.Error(t, history.SetMessages(ctx, []llms.ChatMessage{
		llms.HumanChatMessage{Content: "reject"},
	}))
	messages, err := history.Messages(ctx)
	require.NoError(t, err)
	require.Equal(t, []llms.ChatMessage{llms.HumanChatMessage{Content: "keep this message"}}, messages)
	require.NoError(t, history.SetMessages(ctx, []llms.ChatMessage{
		llms.AIChatMessage{Content: "replacement"},
	}))
	messages, err = history.Messages(ctx)
	require.NoError(t, err)
	require.Equal(t, []llms.ChatMessage{llms.AIChatMessage{Content: "replacement"}}, messages)
}
