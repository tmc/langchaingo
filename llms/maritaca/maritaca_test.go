package maritaca

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/0xDezzy/langchaingo/internal/httprr"
	"github.com/0xDezzy/langchaingo/llms"
)

func newTestClient(t *testing.T, opts ...Option) *LLM {
	t.Helper()

	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "MARITACA_KEY")

	rr := httprr.OpenForTest(t, http.DefaultTransport)

	// Configure with httprr HTTP client
	opts = append([]Option{WithHTTPClient(rr.Client()), WithModel("sabia-2-medium")}, opts...)

	c, err := New(opts...)
	require.NoError(t, err)
	return c
}

func TestGenerateContent(t *testing.T) {
	ctx := context.Background()
	t.Parallel()
	llm := newTestClient(t)

	parts := []llms.ContentPart{
		llms.TextContent{Text: "How many feet are in a nautical mile?"},
	}
	content := []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: parts,
		},
	}

	rsp, err := llm.GenerateContent(ctx, content)

	require.NoError(t, err)

	assert.NotEmpty(t, rsp.Choices)
	c1 := rsp.Choices[0]
	assert.Regexp(t, "feet", strings.ToLower(c1.Content))
}

func TestWithStreaming(t *testing.T) {
	ctx := context.Background()
	t.Parallel()
	llm := newTestClient(t)

	parts := []llms.ContentPart{
		llms.TextContent{Text: "How many feet are in a nautical mile?"},
	}
	content := []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: parts,
		},
	}

	var sb strings.Builder
	rsp, err := llm.GenerateContent(ctx, content,
		llms.WithStreamingFunc(func(_ context.Context, chunk []byte) error {
			sb.Write(chunk)
			return nil
		}))
	require.NoError(t, err)

	assert.NotEmpty(t, rsp.Choices)
	c1 := rsp.Choices[0]
	assert.Regexp(t, "feet", strings.ToLower(c1.Content))
	assert.Regexp(t, "feet", strings.ToLower(sb.String()))
}
