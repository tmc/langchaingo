package googleai

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/llms"
	"google.golang.org/api/option"
)

func newClient(t *testing.T, opts ...option.ClientOption) *GoogleAI {
	t.Helper()

	genaiKey := os.Getenv("GENAI_API_KEY")
	if genaiKey == "" {
		t.Skip("GENAI_API_KEY not set")
		return nil
	}

	opts = append(opts, option.WithAPIKey(genaiKey))

	llm, err := New(context.Background(), opts...)
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

	rsp, err := llm.GenerateContent(context.Background(), parts)
	require.NoError(t, err)

	assert.NotEmpty(t, rsp.Choices)
	c1 := rsp.Choices[0]
	assert.Regexp(t, "dog|canid|canine", strings.ToLower(c1.Content))
}

func TestMultiContentImage(t *testing.T) {
	t.Parallel()

	llm := newClient(t)

	parts := []llms.ContentPart{
		llms.ImageURLContent{URL: "https://github.com/tmc/langchaingo/blob/main/docs/static/img/parrot-icon.png?raw=true"},
		llms.TextContent{Text: "describe this image in detail"},
	}

	rsp, err := llm.GenerateContent(context.Background(), parts, llms.WithModel("gemini-pro-vision"))
	require.NoError(t, err)

	assert.NotEmpty(t, rsp.Choices)
	c1 := rsp.Choices[0]
	assert.Regexp(t, "parrot", strings.ToLower(c1.Content))
}
