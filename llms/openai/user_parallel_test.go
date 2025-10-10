package openai

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestWithUserAndParallelToolCallsOptions verifies that WithUser and WithParallelToolCalls
// options are correctly applied to the client.
func TestWithUserAndParallelToolCallsOptions(t *testing.T) {
	t.Parallel()

	// Test WithUser option
	t.Run("WithUser", func(t *testing.T) {
		llm, err := New(
			WithToken("test-token"),
			WithModel("gpt-4"),
			WithUser("test-user-123"),
		)
		require.NoError(t, err)
		require.NotNil(t, llm)
		
		// Verify the user is set on the client
		assert.Equal(t, "test-user-123", llm.client.User)
	})

	// Test WithParallelToolCalls option
	t.Run("WithParallelToolCalls", func(t *testing.T) {
		llm, err := New(
			WithToken("test-token"),
			WithModel("gpt-4"),
			WithParallelToolCalls(true),
		)
		require.NoError(t, err)
		require.NotNil(t, llm)
		
		// Verify ParallelToolCalls is set on the client
		assert.True(t, llm.client.ParallelToolCalls)
	})

	// Test both options together
	t.Run("BothOptions", func(t *testing.T) {
		llm, err := New(
			WithToken("test-token"),
			WithModel("gpt-4"),
			WithUser("test-user-456"),
			WithParallelToolCalls(false),
		)
		require.NoError(t, err)
		require.NotNil(t, llm)
		
		// Verify both are set correctly
		assert.Equal(t, "test-user-456", llm.client.User)
		assert.False(t, llm.client.ParallelToolCalls)
	})
}

