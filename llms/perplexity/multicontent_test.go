package perplexity

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/schema"
)

func newTestClient(t *testing.T, opts ...Option) llms.Model {
	t.Helper()
	if perplexityKey := os.Getenv("PERPLEXITY_API_KEY"); perplexityKey == "" {
		t.Skip("PERPLEXITY_API_KEY not set")
		return nil
	}

	llm, err := New(opts...)
	require.NoError(t, err)
	return llm
}

func TestMultiContentText(t *testing.T) {
	t.Parallel()
	llm := newTestClient(t)

	parts := []llms.ContentPart{
		llms.TextPart("I'm a pomeranian"),
		llms.TextPart("What kind of mammal am I?"),
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
	llm := newTestClient(t)

	content := []llms.MessageContent{
		{
			Role:  schema.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextPart("Name some countries")},
		},
		{
			Role:  schema.ChatMessageTypeAI,
			Parts: []llms.ContentPart{llms.TextPart("Spain and Lesotho")},
		},
		{
			Role:  schema.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextPart("Which if these is larger?")},
		},
	}

	rsp, err := llm.GenerateContent(context.Background(), content)
	require.NoError(t, err)

	assert.NotEmpty(t, rsp.Choices)
	c1 := rsp.Choices[0]
	assert.Regexp(t, "spain.*larger", strings.ToLower(c1.Content))
}

func TestWithStreaming(t *testing.T) {
	t.Parallel()
	llm := newTestClient(t)

	parts := []llms.ContentPart{
		llms.TextPart("I'm a pomeranian"),
		llms.TextPart("Tell me more about my taxonomy"),
	}
	content := []llms.MessageContent{
		{
			Role:  schema.ChatMessageTypeHuman,
			Parts: parts,
		},
	}

	var sb strings.Builder
	rsp, err := llm.GenerateContent(context.Background(), content,
		llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
			sb.Write(chunk)
			return nil
		}))

	require.NoError(t, err)

	assert.NotEmpty(t, rsp.Choices)
	c1 := rsp.Choices[0]
	assert.Regexp(t, "dog|canid", strings.ToLower(c1.Content))
	assert.Regexp(t, "dog|canid", strings.ToLower(sb.String()))
}

//nolint:lll

func showResponse(rsp any) string { //nolint:golint,unused
	b, err := json.MarshalIndent(rsp, "", "  ")
	if err != nil {
		panic(err)
	}
	return string(b)
}
