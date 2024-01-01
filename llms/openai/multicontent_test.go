package openai

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/llms"
)

func newChatClient(t *testing.T, opts ...Option) *Chat {
	t.Helper()
	if openaiKey := os.Getenv("OPENAI_API_KEY"); openaiKey == "" {
		t.Skip("OPENAI_API_KEY not set")
		return nil
	}

	llm, err := NewChat(opts...)
	require.NoError(t, err)
	return llm
}

func TestMultiContentText(t *testing.T) {
	t.Parallel()

	llm := newChatClient(t)

	parts := []llms.ContentPart{
		llms.TextContent{"I'm a pomeranian"},
		llms.TextContent{"What kind of mammal am I?"},
	}

	rsp, err := llm.GenerateContent(context.Background(), parts)
	require.NoError(t, err)

	assert.GreaterOrEqual(t, len(rsp.Choices), 1)
	c1 := rsp.Choices[0]
	assert.Regexp(t, "dog", strings.ToLower(c1.Content))
}

// TODO: add test for image URL, for stability use some image on our own
// site? maybe parrot-icon.png?
// then add documentation all around and send
// initial PR.
