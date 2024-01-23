package vertex

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

func newClient(t *testing.T) *Vertex {
	t.Helper()

	project := os.Getenv("VERTEX_PROJECT")
	if project == "" {
		t.Skip("VERTEX_PROJECT not set")
		return nil
	}
	location := os.Getenv("VERTEX_LOCATION")
	if location == "" {
		location = "us-central1"
	}

	llm, err := NewVertex(context.Background(), WithCloudProject(project), WithCloudLocation(location))
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
	assert.Regexp(t, "(?i)dog|canid|canine", c1.Content)
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

func TestEmbeddings(t *testing.T) {
	t.Parallel()
	llm := newClient(t)

	texts := []string{"foo", "parrot"}
	res, err := llm.CreateEmbedding(context.Background(), texts)
	require.NoError(t, err)

	fmt.Println(res)
	assert.Equal(t, len(texts), len(res))
	assert.NotEmpty(t, res[0])
	assert.NotEmpty(t, res[1])
}
