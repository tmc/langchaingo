package ollama

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/schema"
)

func newChatClient(t *testing.T) *Chat {
	t.Helper()
	var ollamaModel string
	if ollamaModel = os.Getenv("OLLAMA_TEST_MODEL"); ollamaModel == "" {
		t.Skip("OLLAMA_TEST_MODEL not set")
		return nil
	}

	c, err := NewChat(WithModel(ollamaModel))
	require.NoError(t, err)
	return c
}

//nolint:all
func TestChatBasic(t *testing.T) {
	t.Parallel()

	llm := newChatClient(t)

	resp, err := llm.Call(context.Background(), []schema.ChatMessage{
		schema.SystemChatMessage{Content: "You are producing poems in Spanish."},
		schema.HumanChatMessage{Content: "Write a very short poem about Donald Knuth"},
	})
	require.NoError(t, err)
	assert.Regexp(t, "programa|comput|algoritm|libro", strings.ToLower(resp.Content))
}

//nolint:all
func TestGenerateContent(t *testing.T) {
	t.Parallel()

	llm := newChatClient(t)

	parts := []llms.ContentPart{
		llms.TextContent{Text: "How many feet are in a nautical mile?"},
	}
	content := []llms.MessageContent{
		{
			Role:  schema.ChatMessageTypeHuman,
			Parts: parts,
		},
	}

	rsp, err := llm.GenerateContent(context.Background(), content)
	require.NoError(t, err)

	assert.NotEmpty(t, rsp.Choices)
	c1 := rsp.Choices[0]
	fmt.Println(c1)
	assert.Regexp(t, "feet", strings.ToLower(c1.Content))
}
