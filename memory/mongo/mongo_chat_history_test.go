package mongo

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/log"
	"github.com/testcontainers/testcontainers-go/modules/mongodb"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/tmc/langchaingo/internal/testutil/testctr"
	"github.com/tmc/langchaingo/llms"
)

func runTestContainer(t *testing.T) string {
	t.Helper()

	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx := context.Background()

	mongoContainer, err := mongodb.Run(
		ctx,
		"mongo:7.0.8",
		mongodb.WithUsername("test"),
		mongodb.WithPassword("test"),
		testcontainers.WithLogger(log.TestLogger(t)),
		testcontainers.WithWaitStrategy(
			wait.ForAll(
				wait.ForLog("Waiting for connections").
					WithStartupTimeout(60*time.Second),
				wait.ForListeningPort("27017/tcp").
					WithStartupTimeout(60*time.Second),
			)),
	)
	if err != nil && strings.Contains(err.Error(), "Cannot connect to the Docker daemon") {
		t.Skip("Docker not available")
	}
	require.NoError(t, err)

	t.Cleanup(func() {
		if err := mongoContainer.Terminate(context.Background()); err != nil {
			t.Logf("Failed to terminate mongo container: %v", err)
		}
	})

	url, err := mongoContainer.ConnectionString(ctx)
	require.NoError(t, err)

	// Give the container a moment to fully initialize
	time.Sleep(2 * time.Second)

	return url
}

func TestMongoDBChatMessageHistory(t *testing.T) {
	t.Parallel()
	testctr.SkipIfDockerNotAvailable(t)
	ctx := context.Background()

	url := runTestContainer(t)
	_, err := NewMongoDBChatMessageHistory(ctx, WithSessionID("test"))
	assert.Equal(t, errMongoInvalidURL, err)

	_, err = NewMongoDBChatMessageHistory(ctx, WithConnectionURL(url))
	assert.Equal(t, errMongoInvalidSessionID, err)

	history, err := NewMongoDBChatMessageHistory(ctx, WithConnectionURL(url), WithSessionID("testSessionXX"))
	require.NoError(t, err)

	err = history.AddAIMessage(ctx, "Hi")
	require.NoError(t, err)

	err = history.AddUserMessage(ctx, "Hello")
	require.NoError(t, err)

	messages, err := history.Messages(ctx)
	require.NoError(t, err)

	assert.Len(t, messages, 2)
	assert.Equal(t, llms.ChatMessageTypeAI, messages[0].GetType())
	assert.Equal(t, "Hi", messages[0].GetContent())
	assert.Equal(t, llms.ChatMessageTypeHuman, messages[1].GetType())
	assert.Equal(t, "Hello", messages[1].GetContent())
	t.Cleanup(func() {
		if err := history.Clear(context.Background()); err != nil {
			t.Logf("Failed to clear mongo history: %v", err)
		}
	})
}
