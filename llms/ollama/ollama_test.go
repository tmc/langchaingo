package ollama

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/schema"
)

func newChatClient(t *testing.T) *Chat {
	t.Helper()
	var ollamaModel string
	if ollamaModel = os.Getenv("OLLAMA_TEST_MODEL"); ollamaModel == "" {
		t.Skip("OLLAMA_TEST_MODEL not set")
		return nil
	}

	c, err := NewChat(WithLLMOptions(WithModel(ollamaModel)))
	require.NoError(t, err)
	return c
}

func TestChatBasic(t *testing.T) {
	t.Parallel()

	llm := newChatClient(t)

	resp, err := llm.Call(context.Background(), []schema.ChatMessage{
		schema.SystemChatMessage{Content: "You are producing poems in Spanish."},
		schema.HumanChatMessage{Content: "Write a very short poem about Donald Knuth"},
	})
	require.NoError(t, err)
	fmt.Println(resp.Content)
	assert.Regexp(t, "programa|comput|algoritm|libro", strings.ToLower(resp.Content))
}
