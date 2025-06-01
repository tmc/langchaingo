package mistral

import (
	"errors"
	"testing"

	"github.com/tmc/langchaingo/llms"
)

func TestMapError(t *testing.T) { //nolint:funlen // comprehensive error mapping test
	tests := []struct {
		name         string
		err          error
		expectedCode llms.ErrorCode
		expectedMsg  string
	}{
		{
			name:         "nil error",
			err:          nil,
			expectedCode: "",
			expectedMsg:  "",
		},
		{
			name:         "authentication error - invalid api key",
			err:          errors.New("invalid api key provided"),
			expectedCode: llms.ErrCodeAuthentication,
			expectedMsg:  "Invalid or missing API key",
		},
		{
			name:         "authentication error - unauthorized",
			err:          errors.New("401 Unauthorized"),
			expectedCode: llms.ErrCodeAuthentication,
			expectedMsg:  "Invalid or missing API key",
		},
		{
			name:         "rate limit error - too many requests",
			err:          errors.New("too many requests"),
			expectedCode: llms.ErrCodeRateLimit,
			expectedMsg:  "Rate limit exceeded",
		},
		{
			name:         "rate limit error - 429",
			err:          errors.New("Error 429: Rate limit hit"),
			expectedCode: llms.ErrCodeRateLimit,
			expectedMsg:  "Rate limit exceeded",
		},
		{
			name:         "model not found error",
			err:          errors.New("model not found: mistral-unknown"),
			expectedCode: llms.ErrCodeResourceNotFound,
			expectedMsg:  "Model not found",
		},
		{
			name:         "invalid model error",
			err:          errors.New("invalid model specified"),
			expectedCode: llms.ErrCodeResourceNotFound,
			expectedMsg:  "Model not found",
		},
		{
			name:         "token limit error - context length",
			err:          errors.New("context length exceeded"),
			expectedCode: llms.ErrCodeTokenLimit,
			expectedMsg:  "Token limit exceeded",
		},
		{
			name:         "token limit error - too many tokens",
			err:          errors.New("too many tokens in request"),
			expectedCode: llms.ErrCodeTokenLimit,
			expectedMsg:  "Token limit exceeded",
		},
		{
			name:         "token limit error - max_tokens",
			err:          errors.New("max_tokens parameter is too high"),
			expectedCode: llms.ErrCodeTokenLimit,
			expectedMsg:  "Token limit exceeded",
		},
		{
			name:         "invalid request error - 400",
			err:          errors.New("400 Bad Request"),
			expectedCode: llms.ErrCodeInvalidRequest,
			expectedMsg:  "Invalid request",
		},
		{
			name:         "invalid request error",
			err:          errors.New("invalid request format"),
			expectedCode: llms.ErrCodeInvalidRequest,
			expectedMsg:  "Invalid request",
		},
		{
			name:         "quota exceeded error",
			err:          errors.New("quota exceeded for this API key"),
			expectedCode: llms.ErrCodeQuotaExceeded,
			expectedMsg:  "API quota exceeded",
		},
		{
			name:         "insufficient quota error",
			err:          errors.New("insufficient_quota"),
			expectedCode: llms.ErrCodeQuotaExceeded,
			expectedMsg:  "API quota exceeded",
		},
		{
			name:         "service unavailable error - 503",
			err:          errors.New("503 Service Unavailable"),
			expectedCode: llms.ErrCodeProviderUnavailable,
			expectedMsg:  "Mistral service temporarily unavailable",
		},
		{
			name:         "service unavailable error",
			err:          errors.New("service unavailable, please retry"),
			expectedCode: llms.ErrCodeProviderUnavailable,
			expectedMsg:  "Mistral service temporarily unavailable",
		},
		{
			name:         "internal error - 500",
			err:          errors.New("500 Internal Server Error"),
			expectedCode: llms.ErrCodeProviderUnavailable,
			expectedMsg:  "Mistral service error",
		},
		{
			name:         "internal error",
			err:          errors.New("internal error occurred"),
			expectedCode: llms.ErrCodeProviderUnavailable,
			expectedMsg:  "Mistral service error",
		},
		{
			name:         "unknown error",
			err:          errors.New("some unknown error"),
			expectedCode: llms.ErrCodeUnknown,
			expectedMsg:  "",
		},
		{
			name:         "case insensitive matching",
			err:          errors.New("INVALID API KEY"),
			expectedCode: llms.ErrCodeAuthentication,
			expectedMsg:  "Invalid or missing API key",
		},
		{
			name:         "partial match in longer message",
			err:          errors.New("Request failed with error: rate limit exceeded, please wait"),
			expectedCode: llms.ErrCodeRateLimit,
			expectedMsg:  "Rate limit exceeded",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MapError(tt.err)

			if tt.err == nil {
				if result != nil {
					t.Errorf("MapError() = %v, want nil", result)
				}
				return
			}

			llmErr, ok := result.(*llms.Error)
			if !ok && tt.expectedCode != llms.ErrCodeUnknown {
				t.Errorf("MapError() did not return *llms.Error")
				return
			}

			if ok {
				if llmErr.Code != tt.expectedCode {
					t.Errorf("MapError() Code = %v, want %v", llmErr.Code, tt.expectedCode)
				}
				if tt.expectedMsg != "" && llmErr.Message != tt.expectedMsg {
					t.Errorf("MapError() Message = %v, want %v", llmErr.Message, tt.expectedMsg)
				}
				if llmErr.Provider != "mistral" {
					t.Errorf("MapError() Provider = %v, want %v", llmErr.Provider, "mistral")
				}
				// Check if the original error is wrapped
				if !errors.Is(result, tt.err) && tt.expectedCode != llms.ErrCodeUnknown {
					t.Errorf("MapError() did not wrap original error correctly")
				}
			}
		})
	}
}

func TestErrorMappingCompleteness(t *testing.T) {
	// This test verifies that all error patterns in mistralErrorMappings are valid
	for i, mapping := range mistralErrorMappings {
		if len(mapping.patterns) == 0 {
			t.Errorf("Error mapping at index %d has no patterns", i)
		}
		if mapping.code == "" {
			t.Errorf("Error mapping at index %d has no error code", i)
		}
		if mapping.message == "" {
			t.Errorf("Error mapping at index %d has no message", i)
		}
	}
}
