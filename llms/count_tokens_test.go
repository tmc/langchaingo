package llms

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCountTokens(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		model         string
		text          string
		expectedCount int
	}{
		{
			name:          "gpt-3.5-turbo",
			model:         "gpt-3.5-turbo",
			text:          "test for counting tokens",
			expectedCount: 4,
		},
		{
			name:          "gpt-4o",
			model:         "gpt-4o",
			text:          "test for counting tokens",
			expectedCount: 4,
		},
		{
			name:          "gpt-4o-mini",
			model:         "gpt-4o-mini",
			text:          "test for counting tokens",
			expectedCount: 4,
		},
		{
			name:          "gpt-4.1",
			model:         "gpt-4.1",
			text:          "test for counting tokens",
			expectedCount: 4,
		},
		{
			name:          "gpt-4.1-mini",
			model:         "gpt-4.1-mini",
			text:          "test for counting tokens",
			expectedCount: 4,
		},
		{
			name:          "gpt-4.1-nano",
			model:         "gpt-4.1-nano",
			text:          "test for counting tokens",
			expectedCount: 4,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			numTokens := CountTokens(tt.model, tt.text)
			assert.Equal(t, tt.expectedCount, numTokens)
		})
	}
}
