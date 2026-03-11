package llms

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCountTokens(t *testing.T) {
	t.Parallel()
	numTokens := CountTokens("gpt-3.5-turbo", "test for counting tokens")
	expectedNumTokens := 4
	assert.Equal(t, expectedNumTokens, numTokens)
}

func TestGetModelContextSize(t *testing.T) {
	t.Parallel()
	tests := []struct {
		model        string
		expectedSize int
	}{
		// GPT-3.5 models
		{"gpt-3.5-turbo", 16385},
		{"gpt-3.5-turbo-16k", 16385},
		{"gpt-3.5-turbo-0125", 16385},
		{"gpt-3.5-turbo-1106", 16385},
		// GPT-4 models
		{"gpt-4", 8192},
		{"gpt-4-32k", 32768},
		{"gpt-4-0613", 8192},
		{"gpt-4-32k-0613", 32768},
		// GPT-4 Turbo models
		{"gpt-4-turbo", 128000},
		{"gpt-4-turbo-preview", 128000},
		{"gpt-4-turbo-2024-04-09", 128000},
		{"gpt-4-1106-preview", 128000},
		{"gpt-4-0125-preview", 128000},
		// GPT-4o models
		{"gpt-4o", 128000},
		{"gpt-4o-2024-05-13", 128000},
		{"gpt-4o-2024-08-06", 128000},
		{"gpt-4o-mini", 128000},
		{"gpt-4o-mini-2024-07-18", 128000},
		// Legacy models
		{"text-davinci-003", 4097},
		{"text-curie-001", 2048},
		{"text-babbage-001", 2048},
		{"text-ada-001", 2048},
		{"code-davinci-002", 8000},
		{"code-cushman-001", 2048},
		// Unknown model should return default
		{"unknown-model", 2048},
	}

	for _, tt := range tests {
		t.Run(tt.model, func(t *testing.T) {
			size := GetModelContextSize(tt.model)
			assert.Equal(t, tt.expectedSize, size, "Context size for model %s", tt.model)
		})
	}
}
