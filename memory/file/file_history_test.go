package file

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/llms"
)

func TestFileChatMessageHistory(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "file_chat_history_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	filePath := filepath.Join(tempDir, "chat_history.json")

	history, err := NewFileChatMessageHistory(
		WithFilePath(filePath),
		WithCreateDirIfNotExist(true),
	)
	require.NoError(t, err)

	// Test adding messages
	ctx := context.Background()
	err = history.AddUserMessage(ctx, "Hello")
	require.NoError(t, err)
	err = history.AddAIMessage(ctx, "Hi there! How can I help you?")
	require.NoError(t, err)

	// Test retrieving messages
	messages, err := history.Messages(ctx)
	require.NoError(t, err)
	assert.Equal(t, 2, len(messages))
	assert.Equal(t, "Hello", messages[0].GetContent())
	assert.Equal(t, "Hi there! How can I help you?", messages[1].GetContent())

	// Test adding custom message
	err = history.AddMessage(ctx, llms.SystemChatMessage{Content: "This is a system message"})
	require.NoError(t, err)
	messages, err = history.Messages(ctx)
	require.NoError(t, err)
	assert.Equal(t, 3, len(messages))
	assert.Equal(t, "This is a system message", messages[2].GetContent())

	// Test setting messages
	newMessages := []llms.ChatMessage{
		llms.HumanChatMessage{Content: "New message 1"},
		llms.AIChatMessage{Content: "New message 2"},
	}
	err = history.SetMessages(ctx, newMessages)
	require.NoError(t, err)
	messages, err = history.Messages(ctx)
	require.NoError(t, err)
	assert.Equal(t, 2, len(messages))
	assert.Equal(t, "New message 1", messages[0].GetContent())
	assert.Equal(t, "New message 2", messages[1].GetContent())

	// Test clearing messages
	err = history.Clear(ctx)
	require.NoError(t, err)
	messages, err = history.Messages(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, len(messages))

	// Test file persistence
	err = history.AddUserMessage(ctx, "Test persistence")
	require.NoError(t, err)

	// Create new instance, should be able to read previously saved messages
	newHistory, err := NewFileChatMessageHistory(
		WithFilePath(filePath),
	)
	require.NoError(t, err)
	messages, err = newHistory.Messages(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, len(messages))
	assert.Equal(t, "Test persistence", messages[0].GetContent())
}

func TestFileChatHistoryWithNestedDirectory(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "file_chat_history_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	nestedDir := filepath.Join(tempDir, "nested", "dir")
	filePath := filepath.Join(nestedDir, "chat_history.json")

	history, err := NewFileChatMessageHistory(
		WithFilePath(filePath),
		WithCreateDirIfNotExist(true),
	)
	require.NoError(t, err)

	// Verify directory was created
	_, err = os.Stat(nestedDir)
	assert.NoError(t, err)

	// Test adding message
	ctx := context.Background()
	err = history.AddUserMessage(ctx, "Test nested directory")
	require.NoError(t, err)

	// Verify message was saved
	messages, err := history.Messages(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, len(messages))
	assert.Equal(t, "Test nested directory", messages[0].GetContent())
}
