package prompts

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStringPromptValueString(t *testing.T) {
	t.Parallel()

	spv := NewStringPromptValue("")
	str := spv.String()
	assert.Empty(t, str)

	spv = NewStringPromptValue("test")
	str = spv.String()
	assert.Equal(t, "test", str)
}

func TestStringPromptValueMessages(t *testing.T) {
	t.Parallel()

	spv := NewStringPromptValue("")
	msgs := spv.Messages()
	require.Len(t, msgs, 1)

	spv = NewStringPromptValue("test")
	msgs = spv.Messages()
	require.Len(t, msgs, 1)
	assert.Equal(t, "test", msgs[0].GetText())
}
