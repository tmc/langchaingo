package githubmodels

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/llms"
)

func skipIfNoGitHubToken(t *testing.T) string {
	t.Helper()
	token := os.Getenv(tokenEnvVarName)
	if token == "" {
		t.Skip("Skipping test because GITHUB_TOKEN environment variable is not set")
	}
	return token
}

func TestNew(t *testing.T) {
	t.Run("with token from env", func(t *testing.T) {
		token := skipIfNoGitHubToken(t)
		os.Setenv(tokenEnvVarName, token)
		defer os.Unsetenv(tokenEnvVarName)

		llm, err := New()
		require.NoError(t, err)
		require.NotNil(t, llm)
	})

	t.Run("with explicit token", func(t *testing.T) {
		token := skipIfNoGitHubToken(t)
		llm, err := New(WithToken(token))
		require.NoError(t, err)
		require.NotNil(t, llm)
	})

	t.Run("no token", func(t *testing.T) {
		os.Unsetenv(tokenEnvVarName)
		llm, err := New()
		require.Error(t, err)
		assert.Equal(t, ErrMissingToken, err)
		assert.Nil(t, llm)
	})
}

func TestCall(t *testing.T) {
	token := skipIfNoGitHubToken(t)
	llm, err := New(WithToken(token))
	require.NoError(t, err)

	response, err := llm.Call(context.Background(), "What is the capital of France?")
	require.NoError(t, err)
	assert.Contains(t, response, "Paris")
}

func TestGenerateContent(t *testing.T) {
	token := skipIfNoGitHubToken(t)
	llm, err := New(WithToken(token))
	require.NoError(t, err)

	messages := []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeSystem,
			Parts: []llms.ContentPart{llms.TextContent{Text: "You are a helpful assistant."}},
		},
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextContent{Text: "What is the capital of France?"}},
		},
	}

	response, err := llm.GenerateContent(context.Background(), messages)
	require.NoError(t, err)
	assert.NotEmpty(t, response.Choices)
	assert.Contains(t, response.Choices[0].Content, "Paris")
}
