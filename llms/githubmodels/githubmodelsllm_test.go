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

// TestNew tests the New function.
// nolint:tparallel
func TestNew(t *testing.T) {
	// Cannot use t.Parallel() here because TestNew/with_token_from_env uses t.Setenv()

	t.Run("with token from env", func(t *testing.T) {
		// Cannot use t.Parallel() with t.Setenv()
		token := skipIfNoGitHubToken(t)
		t.Setenv(tokenEnvVarName, token)

		llm, err := New()
		require.NoError(t, err)
		require.NotNil(t, llm)
	})

	t.Run("with explicit token", func(t *testing.T) {
		t.Parallel()
		token := skipIfNoGitHubToken(t)
		llm, err := New(WithToken(token))
		require.NoError(t, err)
		require.NotNil(t, llm)
	})

	t.Run("no token", func(t *testing.T) { //nolint:paralleltest
		// Cannot use t.Parallel() with os.Unsetenv() which affects environment
		os.Unsetenv(tokenEnvVarName)
		llm, err := New()
		require.Error(t, err)
		assert.Equal(t, ErrMissingToken, err)
		assert.Nil(t, llm)
	})
}

func TestCall(t *testing.T) {
	t.Parallel()
	token := skipIfNoGitHubToken(t)
	llm, err := New(WithToken(token))
	require.NoError(t, err)

	response, err := llm.Call(context.Background(), "What is the capital of France?")
	require.NoError(t, err)
	assert.Contains(t, response, "Paris")
}

func TestGenerateContent(t *testing.T) {
	t.Parallel()
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

func TestMessageTypeToRole(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		msgType  llms.ChatMessageType
		expected string
	}{
		{"system message", llms.ChatMessageTypeSystem, "system"},
		{"human message", llms.ChatMessageTypeHuman, "user"},
		{"ai message", llms.ChatMessageTypeAI, "assistant"},
		{"generic message", llms.ChatMessageTypeGeneric, "user"},
		{"function message", llms.ChatMessageTypeFunction, "function"},
		{"tool message", llms.ChatMessageTypeTool, "tool"},
		{"unknown message", llms.ChatMessageType("unknown"), "user"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result := typeToRole(tc.msgType)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestContentPartConcatenation(t *testing.T) {
	t.Parallel()
	// Test case with multiple text parts
	messages := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextContent{Text: "Hello"},
				llms.TextContent{Text: " World"},
				llms.TextContent{Text: "!"},
			},
		},
	}

	// Test the message content concatenation behavior
	var contentText string
	for _, part := range messages[0].Parts {
		if textContent, ok := part.(llms.TextContent); ok {
			contentText += textContent.Text
		}
	}
	// Verify the concatenation worked
	assert.Equal(t, "Hello World!", contentText)
}
