package prompts

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStringPromptValueString(t *testing.T) {
	t.Parallel()

	spv := StringPromptValue("")
	str := spv.String()
	assert.Empty(t, str)

	spv = StringPromptValue("test")
	str = spv.String()
	assert.Equal(t, "test", str)
}

func TestStringPromptValueMessages(t *testing.T) {
	t.Parallel()

	spv := StringPromptValue("")
	msgs := spv.Messages()
	require.Len(t, msgs, 1)

	spv = StringPromptValue("test")
	msgs = spv.Messages()
	require.Len(t, msgs, 1)
	assert.Equal(t, "test", msgs[0].GetContent())
}
