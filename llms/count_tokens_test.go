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
