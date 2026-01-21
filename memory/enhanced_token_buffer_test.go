package memory

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/fake"
)

// MockTokenCounter for testing purposes
type MockTokenCounter struct {
	TokensPerMessage int
	TokensPerChar    float64
}

func (mtc *MockTokenCounter) CountTokens(text string) (int, error) {
	return int(float64(len(text)) * mtc.TokensPerChar), nil
}

func (mtc *MockTokenCounter) CountTokensFromMessages(messages []llms.ChatMessage) (int, error) {
	if mtc.TokensPerMessage > 0 {
		return len(messages) * mtc.TokensPerMessage, nil
	}
	
	total := 0
	for _, msg := range messages {
		tokens, err := mtc.CountTokens(msg.GetContent())
		if err != nil {
			return 0, err
		}
		total += tokens + 3 // Add formatting overhead
	}
	return total, nil
}

func TestEnhancedTokenBuffer_BasicOperations(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	buffer := NewEnhancedTokenBuffer(
		WithTokenLimit(100),
		WithTokenCounter(&MockTokenCounter{TokensPerMessage: 10}),
	)

	// Test initial state
	vars := buffer.MemoryVariables(ctx)
	assert.Equal(t, []string{"history"}, vars)

	// Test empty memory
	memory, err := buffer.LoadMemoryVariables(ctx, nil)
	require.NoError(t, err)
	assert.Equal(t, "", memory["history"])

	// Test saving context
	err = buffer.SaveContext(ctx, 
		map[string]any{"input": "Hello"},
		map[string]any{"output": "Hi there!"})
	require.NoError(t, err)

	// Check memory contains the conversation
	memory, err = buffer.LoadMemoryVariables(ctx, nil)
	require.NoError(t, err)
	assert.Contains(t, memory["history"].(string), "Hello")
	assert.Contains(t, memory["history"].(string), "Hi there!")
}

func TestEnhancedTokenBuffer_TokenLimitTrimming(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	buffer := NewEnhancedTokenBuffer(
		WithTokenLimit(25), // Very low limit to force trimming
		WithTokenCounter(&MockTokenCounter{TokensPerMessage: 10}),
		WithMinMessages(0), // Allow trimming all messages if needed
	)

	// Add multiple conversations to exceed limit
	conversations := []struct {
		input  string
		output string
	}{
		{"First message", "First response"},
		{"Second message", "Second response"},
		{"Third message", "Third response"},
		{"Fourth message", "Fourth response"},
	}

	for _, conv := range conversations {
		err := buffer.SaveContext(ctx,
			map[string]any{"input": conv.input},
			map[string]any{"output": conv.output})
		require.NoError(t, err)
	}

	// Get current token count
	tokenCount, err := buffer.GetTokenCount(ctx)
	require.NoError(t, err)
	assert.LessOrEqual(t, tokenCount, 25, "Token count should be within limit")

	// Check that recent messages are preserved
	memory, err := buffer.LoadMemoryVariables(ctx, nil)
	require.NoError(t, err)
	memoryStr := memory["history"].(string)
	assert.Contains(t, memoryStr, "Fourth message", "Most recent message should be preserved")
}

func TestEnhancedTokenBuffer_TrimStrategies(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testCases := []struct {
		name     string
		strategy TrimStrategy
	}{
		{"TrimOldest", TrimOldest},
		{"TrimMiddle", TrimMiddle},
		{"TrimByImportance", TrimByImportance},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			buffer := NewEnhancedTokenBuffer(
				WithTokenLimit(30),
				WithTokenCounter(&MockTokenCounter{TokensPerMessage: 10}),
				WithTrimStrategy(tc.strategy),
				WithMinMessages(0),
			)

			// Add messages that will exceed the limit
			for i := 0; i < 5; i++ {
				err := buffer.SaveContext(ctx,
					map[string]any{"input": fmt.Sprintf("Message %d", i)},
					map[string]any{"output": fmt.Sprintf("Response %d", i)})
				require.NoError(t, err)
			}

			// Verify token count is within limit
			tokenCount, err := buffer.GetTokenCount(ctx)
			require.NoError(t, err)
			assert.LessOrEqual(t, tokenCount, 30)

			// Get the final memory
			memory, err := buffer.LoadMemoryVariables(ctx, nil)
			require.NoError(t, err)
			assert.NotEmpty(t, memory["history"])
		})
	}
}

func TestEnhancedTokenBuffer_PreservePairs(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	buffer := NewEnhancedTokenBuffer(
		WithTokenLimit(30),
		WithTokenCounter(&MockTokenCounter{TokensPerMessage: 10}),
		WithPreservePairs(true),
		WithMinMessages(2),
	)

	// Add conversations
	for i := 0; i < 4; i++ {
		err := buffer.SaveContext(ctx,
			map[string]any{"input": fmt.Sprintf("Question %d", i)},
			map[string]any{"output": fmt.Sprintf("Answer %d", i)})
		require.NoError(t, err)
	}

	// Get messages to verify pairing
	messages, err := buffer.ChatHistory.Messages(ctx)
	require.NoError(t, err)

	// Should have even number of messages (human-AI pairs)
	if len(messages) > 0 {
		assert.Equal(t, 0, len(messages)%2, "Should preserve human-AI pairs")
	}
}

func TestEnhancedTokenBuffer_ReturnMessages(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	buffer := NewEnhancedTokenBuffer(
		WithReturnMessages(true),
		WithTokenLimit(100),
	)

	// Add a conversation
	err := buffer.SaveContext(ctx,
		map[string]any{"input": "Test input"},
		map[string]any{"output": "Test output"})
	require.NoError(t, err)

	// Load memory as messages
	memory, err := buffer.LoadMemoryVariables(ctx, nil)
	require.NoError(t, err)

	messages, ok := memory["history"].([]llms.ChatMessage)
	require.True(t, ok, "Should return messages as slice")
	assert.Len(t, messages, 2) // One human, one AI message
}

func TestEnhancedTokenBuffer_CustomTokenCounter(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	
	// Custom counter that counts characters directly
	customCounter := &MockTokenCounter{TokensPerChar: 1.0}
	
	buffer := NewEnhancedTokenBuffer(
		WithTokenLimit(50),
		WithTokenCounter(customCounter),
	)

	// Add a long message that will exceed character-based limit
	longMessage := "This is a very long message that should exceed the character limit when counting tokens as characters"
	err := buffer.SaveContext(ctx,
		map[string]any{"input": longMessage},
		map[string]any{"output": "Short response"})
	require.NoError(t, err)

	// Verify the custom counter is being used
	tokenCount, err := buffer.GetTokenCount(ctx)
	require.NoError(t, err)
	
	// Should be roughly the length of the messages plus formatting
	expectedApprox := len(longMessage) + len("Short response") + 6 // formatting overhead
	assert.InDelta(t, expectedApprox, tokenCount, float64(expectedApprox)*0.1)
}

func TestEnhancedTokenBuffer_Configuration(t *testing.T) {
	t.Parallel()

	buffer := NewEnhancedTokenBuffer(
		WithTokenLimit(1000),
		WithEncodingModel("gpt-4"),
		WithTrimStrategy(TrimMiddle),
		WithPreservePairs(false),
		WithMinMessages(5),
		WithInputKey("question"),
		WithOutputKey("answer"),
		WithMemoryKey("chat_history"),
		WithHumanPrefix("User"),
		WithAIPrefix("Assistant"),
	)

	// Verify configuration
	assert.Equal(t, 1000, buffer.GetTokenLimit())
	assert.Equal(t, "gpt-4", buffer.EncodingModel)
	assert.Equal(t, TrimMiddle, buffer.GetTrimStrategy())
	assert.False(t, buffer.PreservePairs)
	assert.Equal(t, 5, buffer.MinMessages)
	assert.Equal(t, "question", buffer.InputKey)
	assert.Equal(t, "answer", buffer.OutputKey)
	assert.Equal(t, "chat_history", buffer.MemoryKey)
	assert.Equal(t, "User", buffer.HumanPrefix)
	assert.Equal(t, "Assistant", buffer.AIPrefix)
}

func TestEnhancedTokenBuffer_TikTokenCounter(t *testing.T) {
	t.Parallel()

	counter := &TikTokenCounter{ModelName: "gpt-3.5-turbo"}
	
	// Test basic token counting
	tokens, err := counter.CountTokens("Hello world")
	require.NoError(t, err)
	assert.Greater(t, tokens, 0)
	
	// Test message counting
	messages := []llms.ChatMessage{
		llms.HumanMessage{Content: "Hello"},
		llms.AIMessage{Content: "Hi there!"},
	}
	
	messageTokens, err := counter.CountTokensFromMessages(messages)
	require.NoError(t, err)
	assert.Greater(t, messageTokens, tokens) // Should include formatting overhead
}

func TestEnhancedTokenBuffer_LLMTokenCounter(t *testing.T) {
	t.Parallel()

	llm := fake.NewFakeLLM(fake.WithReturnString("test response"))
	counter := &LLMTokenCounter{LLM: llm, Model: "test-model"}
	
	// Test token counting
	tokens, err := counter.CountTokens("Test message")
	require.NoError(t, err)
	assert.Greater(t, tokens, 0)
}

func TestEnhancedTokenBuffer_Clear(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	buffer := NewEnhancedTokenBuffer()

	// Add some messages
	err := buffer.SaveContext(ctx,
		map[string]any{"input": "Test"},
		map[string]any{"output": "Response"})
	require.NoError(t, err)

	// Verify messages exist
	memory, err := buffer.LoadMemoryVariables(ctx, nil)
	require.NoError(t, err)
	assert.NotEmpty(t, memory["history"])

	// Clear memory
	err = buffer.Clear(ctx)
	require.NoError(t, err)

	// Verify memory is empty
	memory, err = buffer.LoadMemoryVariables(ctx, nil)
	require.NoError(t, err)
	assert.Empty(t, memory["history"])
}

func TestEnhancedTokenBuffer_MinMessagesPreservation(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	buffer := NewEnhancedTokenBuffer(
		WithTokenLimit(1), // Extremely low limit
		WithTokenCounter(&MockTokenCounter{TokensPerMessage: 10}),
		WithMinMessages(2), // But preserve at least 2 messages
	)

	// Add multiple messages
	for i := 0; i < 3; i++ {
		err := buffer.SaveContext(ctx,
			map[string]any{"input": fmt.Sprintf("Input %d", i)},
			map[string]any{"output": fmt.Sprintf("Output %d", i)})
		require.NoError(t, err)
	}

	// Check that minimum messages are preserved despite low token limit
	messages, err := buffer.ChatHistory.Messages(ctx)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(messages), 2, "Should preserve minimum number of messages")
}

// Benchmark tests
func BenchmarkEnhancedTokenBuffer_SaveContext(b *testing.B) {
	ctx := context.Background()
	buffer := NewEnhancedTokenBuffer(
		WithTokenLimit(1000),
		WithTokenCounter(&MockTokenCounter{TokensPerMessage: 10}),
	)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = buffer.SaveContext(ctx,
			map[string]any{"input": "Benchmark input"},
			map[string]any{"output": "Benchmark output"})
	}
}

func BenchmarkEnhancedTokenBuffer_TrimContext(b *testing.B) {
	ctx := context.Background()
	buffer := NewEnhancedTokenBuffer(
		WithTokenLimit(50),
		WithTokenCounter(&MockTokenCounter{TokensPerMessage: 10}),
	)

	// Pre-populate with messages
	for i := 0; i < 20; i++ {
		_ = buffer.SaveContext(ctx,
			map[string]any{"input": fmt.Sprintf("Input %d", i)},
			map[string]any{"output": fmt.Sprintf("Output %d", i)})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = buffer.TrimContext(ctx)
	}
}