package memory

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/pkoukk/tiktoken-go"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/schema"
)

// EnhancedTokenBuffer is an advanced memory buffer that provides sophisticated token management
// with multiple trimming strategies, token counting interfaces, and enhanced configurability.
// This addresses the discussion in https://github.com/tmc/langchaingo/discussions/124
type EnhancedTokenBuffer struct {
	ChatHistory schema.ChatMessageHistory

	// Basic configuration
	ReturnMessages bool
	InputKey       string
	OutputKey      string
	HumanPrefix    string
	AIPrefix       string
	MemoryKey      string

	// Token management configuration
	TokenLimit      int                    // Maximum tokens to keep in memory
	EncodingModel   string                 // Model name for token counting (e.g., "gpt-3.5-turbo")
	TokenCounter    TokenCounter           // Interface for counting tokens
	LLM             llms.Model             // LLM for token counting if TokenCounter not provided
	TrimStrategy    TrimStrategy           // Strategy for trimming messages
	PreservePairs   bool                   // Whether to preserve human-AI message pairs when trimming
	MinMessages     int                    // Minimum number of messages to preserve
}

// TrimStrategy defines how messages should be trimmed when token limit is exceeded.
type TrimStrategy int

const (
	// TrimOldest removes the oldest messages first (default behavior)
	TrimOldest TrimStrategy = iota
	// TrimMiddle preserves the first and last few messages, removing from the middle
	TrimMiddle
	// TrimByImportance attempts to preserve more important messages (experimental)
	TrimByImportance
)

// TokenCounter defines the interface for counting tokens in text.
// This interface allows users to plug in their own token counting implementations.
type TokenCounter interface {
	CountTokens(text string) (int, error)
	CountTokensFromMessages(messages []llms.ChatMessage) (int, error)
}

// TikTokenCounter provides integration with tiktoken tokenizers.
type TikTokenCounter struct {
	ModelName string
	encoder   *tiktoken.Tiktoken
}

// CountTokens counts tokens using model-specific encoding.
func (tc *TikTokenCounter) CountTokens(text string) (int, error) {
	if tc.encoder == nil {
		encoder, err := tiktoken.EncodingForModel(tc.ModelName)
		if err != nil {
			// Fall back to approximation if model not supported
			return tc.approximateTokenCount(text), nil
		}
		tc.encoder = encoder
	}
	
	tokens := tc.encoder.Encode(text, nil, nil)
	return len(tokens), nil
}

// approximateTokenCount provides a fallback token count estimation
func (tc *TikTokenCounter) approximateTokenCount(text string) int {
	switch {
	case strings.Contains(tc.ModelName, "gpt-4"):
		return (len(text) + 2) / 3 // GPT-4 is more efficient
	case strings.Contains(tc.ModelName, "gpt-3.5"):
		return (len(text) + 3) / 4 // Standard GPT-3.5 ratio
	default:
		return (len(text) + 3) / 4 // Conservative estimate
	}
}

// CountTokensFromMessages counts tokens from chat messages with proper formatting overhead.
func (tc *TikTokenCounter) CountTokensFromMessages(messages []llms.ChatMessage) (int, error) {
	if tc.encoder == nil {
		encoder, err := tiktoken.EncodingForModel(tc.ModelName)
		if err != nil {
			// Fall back to approximation
			return tc.approximateTokenCountFromMessages(messages), nil
		}
		tc.encoder = encoder
	}
	
	// Use tiktoken's built-in chat completion token counting for supported models
	if strings.Contains(tc.ModelName, "gpt-3.5") || strings.Contains(tc.ModelName, "gpt-4") {
		// Convert to tiktoken format
		var chatMessages []tiktoken.ChatCompletionMessage
		for _, msg := range messages {
			role := "user"
			switch msg.GetType() {
			case llms.ChatMessageTypeAI:
				role = "assistant"
			case llms.ChatMessageTypeSystem:
				role = "system"
			case llms.ChatMessageTypeHuman:
				role = "user"
			}
			
			chatMessages = append(chatMessages, tiktoken.ChatCompletionMessage{
				Role:    role,
				Content: msg.GetContent(),
			})
		}
		
		return tiktoken.NumTokensFromMessages(chatMessages, tc.ModelName)
	}
	
	// Fall back to manual counting for other models
	return tc.approximateTokenCountFromMessages(messages), nil
}

// approximateTokenCountFromMessages provides fallback token counting
func (tc *TikTokenCounter) approximateTokenCountFromMessages(messages []llms.ChatMessage) int {
	total := 0
	
	// Add conversation setup tokens
	total += 3 // Base conversation formatting
	
	for _, msg := range messages {
		// Count content tokens
		contentTokens := tc.approximateTokenCount(msg.GetContent())
		total += contentTokens
		
		// Add role and formatting tokens based on model
		switch {
		case strings.Contains(tc.ModelName, "gpt-4"), strings.Contains(tc.ModelName, "gpt-3.5"):
			total += 3 // OpenAI chat format: <|start|>role<|end|>
		default:
			total += 4 // Conservative estimate for other models
		}
	}
	
	total += 3 // Conversation end tokens
	return total
}

// LLMTokenCounter uses an LLM to count tokens (fallback option).
type LLMTokenCounter struct {
	LLM   llms.Model
	Model string
}

// CountTokens uses the LLM's token counting if available.
func (ltc *LLMTokenCounter) CountTokens(text string) (int, error) {
	// Use the LLM's built-in token counting if available
	return llms.CountTokens(ltc.Model, text), nil
}

// CountTokensFromMessages counts tokens from messages using LLM.
func (ltc *LLMTokenCounter) CountTokensFromMessages(messages []llms.ChatMessage) (int, error) {
	bufferString, err := llms.GetBufferString(messages, "Human", "AI")
	if err != nil {
		return 0, err
	}
	return ltc.CountTokens(bufferString)
}

// Statically assert that EnhancedTokenBuffer implements the memory interface.
var _ schema.Memory = &EnhancedTokenBuffer{}

// NewEnhancedTokenBuffer creates a new enhanced token-aware buffer memory.
func NewEnhancedTokenBuffer(options ...EnhancedTokenBufferOption) *EnhancedTokenBuffer {
	return applyEnhancedTokenBufferOptions(options...)
}

// MemoryVariables gets the input key the buffer memory class will load dynamically.
func (etb *EnhancedTokenBuffer) MemoryVariables(context.Context) []string {
	return []string{etb.MemoryKey}
}

// LoadMemoryVariables returns the previous chat messages stored in memory after ensuring
// the token count is within limits.
func (etb *EnhancedTokenBuffer) LoadMemoryVariables(
	ctx context.Context, _ map[string]any,
) (map[string]any, error) {
	// Ensure context is trimmed before loading
	err := etb.TrimContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to trim context: %w", err)
	}

	messages, err := etb.ChatHistory.Messages(ctx)
	if err != nil {
		return nil, err
	}

	if etb.ReturnMessages {
		return map[string]any{
			etb.MemoryKey: messages,
		}, nil
	}

	bufferString, err := schema.GetBufferString(messages, etb.HumanPrefix, etb.AIPrefix)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		etb.MemoryKey: bufferString,
	}, nil
}

// SaveContext saves the context from this conversation turn to buffer.
func (etb *EnhancedTokenBuffer) SaveContext(
	ctx context.Context,
	inputValues map[string]any,
	outputValues map[string]any,
) error {
	input, ok := inputValues[etb.InputKey]
	if !ok {
		return ErrInvalidInputValues
	}

	output, ok := outputValues[etb.OutputKey]
	if !ok {
		return ErrInvalidInputValues
	}

	// Save the new messages
	err := etb.ChatHistory.AddUserMessage(ctx, fmt.Sprintf("%s", input))
	if err != nil {
		return err
	}

	err = etb.ChatHistory.AddAIMessage(ctx, fmt.Sprintf("%s", output))
	if err != nil {
		return err
	}

	// Trim context after adding new messages
	return etb.TrimContext(ctx)
}

// Clear removes all messages from the chat history.
func (etb *EnhancedTokenBuffer) Clear(ctx context.Context) error {
	return etb.ChatHistory.Clear(ctx)
}

// TrimContext removes messages according to the configured strategy until
// the total token count is within limits.
func (etb *EnhancedTokenBuffer) TrimContext(ctx context.Context) error {
	if etb.TokenLimit <= 0 {
		return nil // No limit set
	}

	messages, err := etb.ChatHistory.Messages(ctx)
	if err != nil {
		return err
	}

	if len(messages) <= etb.MinMessages {
		return nil // Don't trim below minimum message count
	}

	// Get token counter
	counter := etb.getTokenCounter()

	// Count current tokens
	tokenCount, err := counter.CountTokensFromMessages(messages)
	if err != nil {
		return fmt.Errorf("failed to count tokens: %w", err)
	}

	// Trim messages if over limit
	if tokenCount > etb.TokenLimit {
		trimmedMessages, err := etb.trimMessagesToLimit(messages, counter)
		if err != nil {
			return err
		}
		
		// Clear and re-add trimmed messages
		err = etb.ChatHistory.Clear(ctx)
		if err != nil {
			return err
		}

		for _, msg := range trimmedMessages {
			switch msg.GetType() {
			case llms.ChatMessageTypeHuman:
				err = etb.ChatHistory.AddUserMessage(ctx, msg.GetContent())
			case llms.ChatMessageTypeAI:
				err = etb.ChatHistory.AddAIMessage(ctx, msg.GetContent())
			case llms.ChatMessageTypeSystem:
				// Handle system messages if the history supports them
				if systemHistory, ok := etb.ChatHistory.(interface {
					AddSystemMessage(ctx context.Context, content string) error
				}); ok {
					err = systemHistory.AddSystemMessage(ctx, msg.GetContent())
				} else {
					// Skip system messages if not supported
					continue
				}
			}
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// trimMessagesToLimit removes messages according to the configured strategy.
func (etb *EnhancedTokenBuffer) trimMessagesToLimit(messages []llms.ChatMessage, counter TokenCounter) ([]llms.ChatMessage, error) {
	if len(messages) == 0 {
		return messages, nil
	}

	switch etb.TrimStrategy {
	case TrimOldest:
		return etb.trimOldest(messages, counter)
	case TrimMiddle:
		return etb.trimMiddle(messages, counter)
	case TrimByImportance:
		return etb.trimByImportance(messages, counter)
	default:
		return etb.trimOldest(messages, counter)
	}
}

// trimOldest removes the oldest messages to fit within the token limit.
func (etb *EnhancedTokenBuffer) trimOldest(messages []llms.ChatMessage, counter TokenCounter) ([]llms.ChatMessage, error) {
	// Start from the end and work backwards to keep the most recent messages
	for i := len(messages); i >= etb.MinMessages; i-- {
		candidateMessages := messages[len(messages)-i:]
		
		// If preserving pairs, ensure we don't break human-AI pairs
		if etb.PreservePairs && len(candidateMessages) > 0 {
			candidateMessages = etb.preserveMessagePairs(candidateMessages)
		}
		
		tokenCount, err := counter.CountTokensFromMessages(candidateMessages)
		if err != nil {
			return nil, err
		}

		if tokenCount <= etb.TokenLimit {
			return candidateMessages, nil
		}
	}

	// If we can't fit within limits, return minimum messages
	if len(messages) >= etb.MinMessages {
		return messages[len(messages)-etb.MinMessages:], nil
	}
	
	return messages, nil
}

// trimMiddle preserves the first and last few messages, removing from the middle.
func (etb *EnhancedTokenBuffer) trimMiddle(messages []llms.ChatMessage, counter TokenCounter) ([]llms.ChatMessage, error) {
	if len(messages) <= etb.MinMessages {
		return messages, nil
	}

	// Keep first 2 and last 2 messages, remove from middle
	keepFirst := min(2, len(messages)/4)
	keepLast := min(2, len(messages)/4)
	
	if keepFirst+keepLast >= len(messages) {
		return messages, nil
	}

	for middleSize := 0; middleSize <= len(messages)-keepFirst-keepLast; middleSize++ {
		candidateMessages := make([]llms.ChatMessage, 0, keepFirst+keepLast+middleSize)
		
		// Add first messages
		candidateMessages = append(candidateMessages, messages[:keepFirst]...)
		
		// Add middle messages (if any)
		if middleSize > 0 {
			middleStart := len(messages) - keepLast - middleSize
			candidateMessages = append(candidateMessages, messages[middleStart:middleStart+middleSize]...)
		}
		
		// Add last messages
		candidateMessages = append(candidateMessages, messages[len(messages)-keepLast:]...)
		
		tokenCount, err := counter.CountTokensFromMessages(candidateMessages)
		if err != nil {
			return nil, err
		}

		if tokenCount <= etb.TokenLimit {
			return candidateMessages, nil
		}
	}

	// Fallback to keeping just first and last messages
	result := make([]llms.ChatMessage, 0, keepFirst+keepLast)
	result = append(result, messages[:keepFirst]...)
	result = append(result, messages[len(messages)-keepLast:]...)
	return result, nil
}

// trimByImportance attempts to preserve more important messages (experimental).
func (etb *EnhancedTokenBuffer) trimByImportance(messages []llms.ChatMessage, counter TokenCounter) ([]llms.ChatMessage, error) {
	// This is a placeholder for more sophisticated importance-based trimming
	// In a real implementation, this could use:
	// - Message length (longer messages might be more important)
	// - Keyword analysis
	// - User-defined importance scores
	// - Recency with decay
	
	// For now, fall back to trimming oldest while preserving recent messages
	return etb.trimOldest(messages, counter)
}

// preserveMessagePairs ensures we don't break human-AI conversation pairs.
func (etb *EnhancedTokenBuffer) preserveMessagePairs(messages []llms.ChatMessage) []llms.ChatMessage {
	if len(messages) == 0 {
		return messages
	}

	// If the first message is from AI and we have more than one message,
	// remove it to start with a human message
	if len(messages) > 1 && messages[0].GetType() == llms.ChatMessageTypeAI {
		messages = messages[1:]
	}

	// If we have an odd number of messages and the last is human,
	// remove the last message to end with an AI response
	if len(messages)%2 == 1 && len(messages) > 1 && 
	   messages[len(messages)-1].GetType() == llms.ChatMessageTypeHuman {
		messages = messages[:len(messages)-1]
	}

	return messages
}

// getTokenCounter returns the appropriate token counter for this buffer.
func (etb *EnhancedTokenBuffer) getTokenCounter() TokenCounter {
	if etb.TokenCounter != nil {
		return etb.TokenCounter
	}

	if etb.EncodingModel != "" {
		return &TikTokenCounter{ModelName: etb.EncodingModel}
	}

	if etb.LLM != nil {
		return &LLMTokenCounter{LLM: etb.LLM, Model: etb.EncodingModel}
	}

	// Fallback to basic tiktoken-style counter
	return &TikTokenCounter{ModelName: "gpt-3.5-turbo"}
}

// GetTokenCount returns the current token count of the memory buffer.
func (etb *EnhancedTokenBuffer) GetTokenCount(ctx context.Context) (int, error) {
	messages, err := etb.ChatHistory.Messages(ctx)
	if err != nil {
		return 0, err
	}

	counter := etb.getTokenCounter()
	return counter.CountTokensFromMessages(messages)
}

// GetMemoryString returns the current memory as a formatted string.
func (etb *EnhancedTokenBuffer) GetMemoryString(ctx context.Context) (string, error) {
	messages, err := etb.ChatHistory.Messages(ctx)
	if err != nil {
		return "", err
	}

	return schema.GetBufferString(messages, etb.HumanPrefix, etb.AIPrefix)
}

// SetTokenLimit updates the token limit for this buffer.
func (etb *EnhancedTokenBuffer) SetTokenLimit(limit int) {
	etb.TokenLimit = limit
}

// GetTokenLimit returns the current token limit.
func (etb *EnhancedTokenBuffer) GetTokenLimit() int {
	return etb.TokenLimit
}

// SetTrimStrategy updates the trimming strategy.
func (etb *EnhancedTokenBuffer) SetTrimStrategy(strategy TrimStrategy) {
	etb.TrimStrategy = strategy
}

// GetTrimStrategy returns the current trimming strategy.
func (etb *EnhancedTokenBuffer) GetTrimStrategy() TrimStrategy {
	return etb.TrimStrategy
}

// Helper function for min operation
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}