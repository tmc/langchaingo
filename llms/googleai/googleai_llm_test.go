package googleai

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/schema"
)

func newClient(t *testing.T) *GoogleAI {
	t.Helper()

	genaiKey := os.Getenv("GENAI_API_KEY")
	if genaiKey == "" {
		t.Skip("GENAI_API_KEY not set")
		return nil
	}
	llm, err := NewGoogleAI(context.Background(), WithAPIKey(genaiKey))
	require.NoError(t, err)
	return llm
}

func TestMultiContentText(t *testing.T) {
	t.Parallel()
	llm := newClient(t)

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
	assert.Regexp(t, "dog|canid|canine", strings.ToLower(c1.Content))
}

func TestMultiContentTextStream(t *testing.T) {
	t.Parallel()
	llm := newClient(t)

	parts := []llms.ContentPart{
		llms.TextContent{Text: "I'm a pomeranian"},
		llms.TextContent{Text: "Tell me more about my taxonomy"},
	}
	content := []llms.MessageContent{
		{
			Role:  schema.ChatMessageTypeHuman,
			Parts: parts,
		},
	}

	var chunks [][]byte
	var sb strings.Builder
	rsp, err := llm.GenerateContent(context.Background(), content,
		llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
			chunks = append(chunks, chunk)
			sb.Write(chunk)
			return nil
		}))

	require.NoError(t, err)

	assert.NotEmpty(t, rsp.Choices)
	// Check that the combined response contains what we expect
	c1 := rsp.Choices[0]
	assert.Regexp(t, "(?i)dog|canid|canine", c1.Content)

	// Check that multiple chunks were received and they also have words
	// we expect.
	assert.GreaterOrEqual(t, len(chunks), 2)
	assert.Regexp(t, "(?i)dog|canid|canine", sb.String())
}

func TestMultiContentTextChatSequence(t *testing.T) {
	t.Parallel()
	llm := newClient(t)

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

	rsp, err := llm.GenerateContent(context.Background(), content, llms.WithModel("gemini-pro"))
	require.NoError(t, err)

	assert.NotEmpty(t, rsp.Choices)
	c1 := rsp.Choices[0]
	assert.Regexp(t, "spain.*larger", strings.ToLower(c1.Content))
}

func TestMultiContentImage(t *testing.T) {
	t.Parallel()
	llm := newClient(t)

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

	rsp, err := llm.GenerateContent(context.Background(), content, llms.WithModel("gemini-pro-vision"))
	require.NoError(t, err)

	assert.NotEmpty(t, rsp.Choices)
	c1 := rsp.Choices[0]
	assert.Regexp(t, "parrot", strings.ToLower(c1.Content))
}

func TestEmbeddings(t *testing.T) {
	t.Parallel()
	llm := newClient(t)

	texts := []string{"foo", "parrot"}
	res, err := llm.CreateEmbedding(context.Background(), texts)
	require.NoError(t, err)

	assert.Equal(t, len(texts), len(res))
	assert.NotEmpty(t, res[0])
	assert.NotEmpty(t, res[1])
}
