package openai

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/schema"
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
		llms.TextContent{Text: "I'm a pomeranian"},
		llms.TextContent{Text: "What kind of mammal am I?"},
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
	assert.Regexp(t, "dog|canid", strings.ToLower(c1.Content))
}

func TestMultiContentTextChatSequence(t *testing.T) {
	t.Parallel()
	llm := newChatClient(t)

	content := []llms.MessageContent{
		{
			Role:  schema.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextContent{Text: "Name some countries"}},
		},
		{
			Role:  schema.ChatMessageTypeAI,
			Parts: []llms.ContentPart{llms.TextContent{Text: "Spain and Lesotho"}},
		},
		{
			Role:  schema.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextContent{Text: "Which if these is larger?"}},
		},
	}

	rsp, err := llm.GenerateContent(context.Background(), content)
	require.NoError(t, err)

	assert.NotEmpty(t, rsp.Choices)
	c1 := rsp.Choices[0]
	fmt.Println(c1)
	assert.Regexp(t, "spain.*larger.*lesotho", strings.ToLower(c1.Content))
}

func TestMultiContentImage(t *testing.T) {
	t.Parallel()

	llm := newChatClient(t, WithModel("gpt-4-vision-preview"))

	parts := []llms.ContentPart{
		llms.ImageURLContent{URL: "https://github.com/tmc/langchaingo/blob/main/docs/static/img/parrot-icon.png?raw=true"},
		llms.TextContent{Text: "describe this image in detail"},
	}
	content := []llms.MessageContent{
		{
			Role:  schema.ChatMessageTypeHuman,
			Parts: parts,
		},
	}

	rsp, err := llm.GenerateContent(context.Background(), content, llms.WithMaxTokens(300))
	require.NoError(t, err)

	assert.NotEmpty(t, rsp.Choices)
	c1 := rsp.Choices[0]
	assert.Regexp(t, "parrot", strings.ToLower(c1.Content))
}

func showResponse(rsp any) string { //nolint:golint,unused
	b, err := json.MarshalIndent(rsp, "", "  ")
	if err != nil {
		panic(err)
	}
	return string(b)
}
