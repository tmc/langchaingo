package anthropic

import (
	"strings"

	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/llms/reasoning"
)

// errorMapping represents a mapping from error patterns to error codes.
type errorMapping struct {
	patterns []string
	code     llms.ErrorCode
	message  string
}

// anthropicErrorMappings defines the error mappings for Anthropic.
var anthropicErrorMappings = []errorMapping{
	{
		patterns: []string{"invalid api key", "authentication failed", "401"},
		code:     llms.ErrCodeAuthentication,
		message:  "Invalid or missing API key",
	},
	{
		patterns: []string{"rate limit", "too many requests", "429"},
		code:     llms.ErrCodeRateLimit,
		message:  "Rate limit exceeded",
	},
	{
		patterns: []string{"model not found", "invalid model"},
		code:     llms.ErrCodeResourceNotFound,
		message:  "Model not found",
	},
	{
		patterns: []string{"maximum tokens", "context window"},
		code:     llms.ErrCodeTokenLimit,
		message:  "Token limit exceeded",
	},
	{
		patterns: []string{"content blocked", "safety violation"},
		code:     llms.ErrCodeContentFilter,
		message:  "Content blocked by safety filter",
	},
	{
		patterns: []string{"credit limit", "quota exceeded"},
		code:     llms.ErrCodeQuotaExceeded,
		message:  "API quota exceeded",
	},
	{
		patterns: []string{"invalid request", "400"},
		code:     llms.ErrCodeInvalidRequest,
		message:  "Invalid request",
	},
	{
		patterns: []string{"overloaded", "503", "service unavailable"},
		code:     llms.ErrCodeProviderUnavailable,
		message:  "Anthropic service temporarily unavailable",
	},
}

// MapError maps Anthropic-specific errors to standardized error codes.
func MapError(err error) error {
	if err == nil {
		return nil
	}

	errStr := strings.ToLower(err.Error())

	// Check each error mapping
	for _, mapping := range anthropicErrorMappings {
		for _, pattern := range mapping.patterns {
			if strings.Contains(errStr, pattern) {
				return llms.NewError(mapping.code, "anthropic", mapping.message).WithCause(err)
			}
		}
	}

	// Use the generic error mapper for unrecognized errors
	mapper := llms.NewErrorMapper("anthropic")
	return mapper.Map(err)
}

// ErrAssistantPrefillUnsupported reports that the model rejects a conversation
// ending with an assistant turn.
type ErrAssistantPrefillUnsupported = reasoning.ErrAssistantPrefillUnsupported

// ErrForcedToolUseWithThinking reports the documented gap that manual (budget)
// thinking accepts only tool_choice "auto" or "none": forcing a tool with "any"
// or a named tool is rejected. Adaptive thinking has no such limit.
type ErrForcedToolUseWithThinking = reasoning.ErrForcedToolUseWithThinking
