package ollama

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/llms"
)

func TestWithContext(t *testing.T) {
	// Test that the WithContext option properly sets the context
	llm, err := New(
		WithModel("llama2"),
		WithContext([]int{1, 2, 3, 4, 5}),
	)
	require.NoError(t, err)

	// Verify the context was set in the options
	expectedContext := []int{1, 2, 3, 4, 5}
	assert.Equal(t, expectedContext, llm.options.context)
}

// TestContextFunctionality is a more comprehensive test that would require
// an actual Ollama server running. This test is commented out but shows
// how context would be used in practice.
/*
func TestContextFunctionality(t *testing.T) {
	// Skip if Ollama is not available
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// First request - should return context
	llm, err := New(WithModel("llama2"))
	require.NoError(t, err)

	ctx := context.Background()
	
	// Make first request
	response1, err := llm.GenerateContent(ctx, []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextPart("Tell me a joke about programming")},
		},
	})
	require.NoError(t, err)
	require.NotEmpty(t, response1.Choices)

	// Extract context from response (this would need to be implemented)
	// contextFromResponse := response1.Context // This field doesn't exist yet
	
	// Second request using context from first request
	llm2, err := New(
		WithModel("llama2"),
		WithContext(contextFromResponse), // Use context from previous response
	)
	require.NoError(t, err)

	response2, err := llm2.GenerateContent(ctx, []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextPart("Tell me another one")},
		},
	})
	require.NoError(t, err)
	require.NotEmpty(t, response2.Choices)

	// The second response should reference the first joke due to context
	// This is hard to test automatically, but the context should help the model
	// understand "another one" refers to another programming joke
}
*/

func TestContextExample(t *testing.T) {
	// Example showing how to use context for conversational memory
	
	// Step 1: Create client without context for first interaction
	llm, err := New(WithModel("llama2"))
	require.NoError(t, err)
	
	// Step 2: Make first request (this would normally get a context back)
	// In a real scenario, you'd extract the context from the response
	// and use it in subsequent requests
	
	// Step 3: Create new client with context for follow-up
	contextFromPreviousResponse := []int{1, 2, 3, 4, 5} // This would come from step 2
	llmWithContext, err := New(
		WithModel("llama2"),
		WithContext(contextFromPreviousResponse),
	)
	require.NoError(t, err)
	
	// Verify context is set
	assert.Equal(t, contextFromPreviousResponse, llmWithContext.options.context)
	
	// This demonstrates the API usage pattern even without a running server
}