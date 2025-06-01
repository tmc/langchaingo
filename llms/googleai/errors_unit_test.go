package googleai

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tmc/langchaingo/llms"
)

func TestMapError(t *testing.T) { //nolint:funlen // comprehensive error mapping test
	t.Parallel()

	tests := []struct {
		name         string
		inputError   error
		expectedCode llms.ErrorCode
		expectedMsg  string
	}{
		{
			name:       "nil error",
			inputError: nil,
		},
		{
			name:         "invalid API key",
			inputError:   errors.New("invalid api key provided"),
			expectedCode: llms.ErrCodeAuthentication,
			expectedMsg:  "Invalid or missing API key",
		},
		{
			name:         "API key not valid",
			inputError:   errors.New("API key not valid"),
			expectedCode: llms.ErrCodeAuthentication,
			expectedMsg:  "Invalid or missing API key",
		},
		{
			name:         "401 authentication error",
			inputError:   errors.New("HTTP 401 unauthorized"),
			expectedCode: llms.ErrCodeAuthentication,
			expectedMsg:  "Invalid or missing API key",
		},
		{
			name:         "quota exceeded",
			inputError:   errors.New("quota exceeded for project"),
			expectedCode: llms.ErrCodeRateLimit,
			expectedMsg:  "Rate limit or quota exceeded",
		},
		{
			name:         "rate limit error",
			inputError:   errors.New("rate limit exceeded"),
			expectedCode: llms.ErrCodeRateLimit,
			expectedMsg:  "Rate limit or quota exceeded",
		},
		{
			name:         "429 rate limit",
			inputError:   errors.New("HTTP 429 too many requests"),
			expectedCode: llms.ErrCodeRateLimit,
			expectedMsg:  "Rate limit or quota exceeded",
		},
		{
			name:         "model not found",
			inputError:   errors.New("model not found: invalid-model"),
			expectedCode: llms.ErrCodeResourceNotFound,
			expectedMsg:  "Model not found",
		},
		{
			name:         "404 not found",
			inputError:   errors.New("HTTP 404 not found"),
			expectedCode: llms.ErrCodeResourceNotFound,
			expectedMsg:  "Model not found",
		},
		{
			name:         "token limit exceeded",
			inputError:   errors.New("token limit exceeded for model"),
			expectedCode: llms.ErrCodeTokenLimit,
			expectedMsg:  "Token limit exceeded",
		},
		{
			name:         "context length exceeded",
			inputError:   errors.New("context length too long"),
			expectedCode: llms.ErrCodeTokenLimit,
			expectedMsg:  "Token limit exceeded",
		},
		{
			name:         "harmful content blocked",
			inputError:   errors.New("harmful content detected"),
			expectedCode: llms.ErrCodeContentFilter,
			expectedMsg:  "Content blocked by safety filter",
		},
		{
			name:         "safety filter triggered",
			inputError:   errors.New("safety threshold exceeded"),
			expectedCode: llms.ErrCodeContentFilter,
			expectedMsg:  "Content blocked by safety filter",
		},
		{
			name:         "billing issue",
			inputError:   errors.New("billing account required"),
			expectedCode: llms.ErrCodeQuotaExceeded,
			expectedMsg:  "Billing quota exceeded",
		},
		{
			name:         "payment required",
			inputError:   errors.New("payment required to continue"),
			expectedCode: llms.ErrCodeQuotaExceeded,
			expectedMsg:  "Billing quota exceeded",
		},
		{
			name:         "invalid request",
			inputError:   errors.New("invalid request format"),
			expectedCode: llms.ErrCodeInvalidRequest,
			expectedMsg:  "Invalid request",
		},
		{
			name:         "400 bad request",
			inputError:   errors.New("HTTP 400 bad request"),
			expectedCode: llms.ErrCodeInvalidRequest,
			expectedMsg:  "Invalid request",
		},
		{
			name:         "service unavailable",
			inputError:   errors.New("service unavailable"),
			expectedCode: llms.ErrCodeProviderUnavailable,
			expectedMsg:  "Google AI service temporarily unavailable",
		},
		{
			name:         "503 service unavailable",
			inputError:   errors.New("HTTP 503 service unavailable"),
			expectedCode: llms.ErrCodeProviderUnavailable,
			expectedMsg:  "Google AI service temporarily unavailable",
		},
		{
			name:         "unrecognized error",
			inputError:   errors.New("some unknown error occurred"),
			expectedCode: llms.ErrCodeUnknown,
			expectedMsg:  "", // Will use generic mapping
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MapError(tt.inputError)

			if tt.inputError == nil {
				assert.Nil(t, result)
				return
			}

			assert.Error(t, result)

			// Check if it's an llms.Error
			if llmsErr, ok := result.(*llms.Error); ok {
				assert.Equal(t, tt.expectedCode, llmsErr.Code)
				assert.Equal(t, "googleai", llmsErr.Provider)
				if tt.expectedMsg != "" {
					assert.Equal(t, tt.expectedMsg, llmsErr.Message)
				}
				assert.Equal(t, tt.inputError, llmsErr.Cause)
			} else if tt.expectedCode != llms.ErrCodeUnknown {
				// For known error codes, we expect an llms.Error
				t.Errorf("Expected llms.Error but got %T", result)
			}
		})
	}
}

func TestMapErrorCaseSensitivity(t *testing.T) {
	t.Parallel()

	// Test that error mapping is case insensitive
	tests := []struct {
		name         string
		inputError   error
		expectedCode llms.ErrorCode
	}{
		{
			name:         "uppercase API key error",
			inputError:   errors.New("INVALID API KEY"),
			expectedCode: llms.ErrCodeAuthentication,
		},
		{
			name:         "mixed case quota error",
			inputError:   errors.New("Quota Exceeded"),
			expectedCode: llms.ErrCodeRateLimit,
		},
		{
			name:         "mixed case model not found",
			inputError:   errors.New("Model Not Found"),
			expectedCode: llms.ErrCodeResourceNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MapError(tt.inputError)
			assert.Error(t, result)

			if llmsErr, ok := result.(*llms.Error); ok {
				assert.Equal(t, tt.expectedCode, llmsErr.Code)
			}
		})
	}
}

func TestErrorMappingStructure(t *testing.T) {
	t.Parallel()

	// Test that all error mappings have required fields
	for i, mapping := range googleAIErrorMappings {
		t.Run(mapping.message, func(t *testing.T) {
			assert.NotEmpty(t, mapping.patterns, "mapping %d should have patterns", i)
			assert.NotEqual(t, llms.ErrorCode(""), mapping.code, "mapping %d should have a valid error code", i)
			assert.NotEmpty(t, mapping.message, "mapping %d should have a message", i)
		})
	}
}

func TestErrorMappingCoverage(t *testing.T) {
	t.Parallel()

	// Ensure we have mappings for all common error scenarios
	expectedErrorCodes := []llms.ErrorCode{
		llms.ErrCodeAuthentication,
		llms.ErrCodeRateLimit,
		llms.ErrCodeResourceNotFound,
		llms.ErrCodeTokenLimit,
		llms.ErrCodeContentFilter,
		llms.ErrCodeQuotaExceeded,
		llms.ErrCodeInvalidRequest,
		llms.ErrCodeProviderUnavailable,
	}

	foundCodes := make(map[llms.ErrorCode]bool)
	for _, mapping := range googleAIErrorMappings {
		foundCodes[mapping.code] = true
	}

	for _, expectedCode := range expectedErrorCodes {
		assert.True(t, foundCodes[expectedCode], "Missing error mapping for code: %v", expectedCode)
	}
}
