package bedrock

import (
	"strings"

	"github.com/0xDezzy/langchaingo/llms"
)

// errorMapping represents a mapping from error patterns to error codes.
type errorMapping struct {
	patterns []string
	code     llms.ErrorCode
	message  string
}

// bedrockErrorMappings defines the error mappings for AWS Bedrock.
var bedrockErrorMappings = []errorMapping{
	{
		patterns: []string{"accessdeniedexception", "unauthorized", "invalid security token"},
		code:     llms.ErrCodeAuthentication,
		message:  "Invalid or missing AWS credentials",
	},
	{
		patterns: []string{"throttlingexception", "toomanyrequestsexception"},
		code:     llms.ErrCodeRateLimit,
		message:  "Request rate limit exceeded",
	},
	{
		patterns: []string{"resourcenotfoundexception", "model not found"},
		code:     llms.ErrCodeResourceNotFound,
		message:  "Model not found or not accessible",
	},
	{
		patterns: []string{"validationexception", "malformed"},
		code:     llms.ErrCodeInvalidRequest,
		message:  "Invalid request parameters",
	},
	{
		patterns: []string{"modeltimeoutexception"},
		code:     llms.ErrCodeTimeout,
		message:  "Model invocation timeout",
	},
	{
		patterns: []string{"serviceexception", "internalservererror", "500"},
		code:     llms.ErrCodeProviderUnavailable,
		message:  "AWS Bedrock service error",
	},
	{
		patterns: []string{"modelnotreadyexception"},
		code:     llms.ErrCodeProviderUnavailable,
		message:  "Model not ready for invocation",
	},
	{
		patterns: []string{"payload size", "token limit"},
		code:     llms.ErrCodeTokenLimit,
		message:  "Input size or token limit exceeded",
	},
}

// MapError maps AWS Bedrock-specific errors to standardized error codes.
func MapError(err error) error {
	if err == nil {
		return nil
	}

	errStr := strings.ToLower(err.Error())

	// Check each error mapping
	for _, mapping := range bedrockErrorMappings {
		for _, pattern := range mapping.patterns {
			if strings.Contains(errStr, pattern) {
				return llms.NewError(mapping.code, "bedrock", mapping.message).WithCause(err)
			}
		}
	}

	// Use the generic error mapper for unrecognized errors
	mapper := llms.NewErrorMapper("bedrock")
	return mapper.Map(err)
}
