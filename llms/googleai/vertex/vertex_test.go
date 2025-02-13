package vertex

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/googleai"
)

func newTestClient(t *testing.T, opts ...googleai.Option) llms.Model {
	t.Helper()
	if googleAiKey := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"); googleAiKey == "" {
		t.Skip("GOOGLE_APPLICATION_CREDENTIALS not set")
		return nil
	}
	if cloudLocation := os.Getenv("GOOGLE_CLOUD_PROJECT_ID"); cloudLocation == "" {
		t.Skip("GOOGLE_CLOUD_PROJECT_ID not set")
		return nil
	}

	llm, err := New(context.Background(), opts...)
	require.NoError(t, err)
	return llm
}

func TestSystemContentText(t *testing.T) {
	t.Parallel()
	llm := newTestClient(t, googleai.WithDefaultModel("gemini-1.5-flash-001"),
		googleai.WithCloudProject("verbeux-dev"),
	)

	system := []llms.ContentPart{
		llms.TextPart("Você é um assistente que responde em português"),
	}
	humanParts := []llms.ContentPart{
		llms.TextPart("I'm a pomeranian"),
		llms.TextPart("What kind of mammal am I?"),
	}
	content := []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeSystem,
			Parts: system,
		},
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: humanParts,
		},
	}

	rsp, err := llm.GenerateContent(context.Background(), content)
	require.NoError(t, err)

	assert.NotEmpty(t, rsp.Choices)
	c1 := rsp.Choices[0]
	assert.Regexp(t, "cão|cães", strings.ToLower(c1.Content))
}

func TestSystemContentTextMoreThanOne(t *testing.T) {
	t.Parallel()
	llm := newTestClient(t, googleai.WithDefaultModel("gemini-1.5-flash-001"),
		googleai.WithCloudProject("verbeux-dev"),
	)

	systemParts := []llms.ContentPart{
		llms.TextPart("Você é um assistente que responde em português"),
	}
	humanParts := []llms.ContentPart{
		llms.TextPart("I'm a pomeranian"),
		llms.TextPart("What kind of mammal am I?"),
	}
	systemPartsTwo := []llms.ContentPart{
		llms.TextPart("Você é um assistente que responde em português"),
	}
	content := []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeSystem,
			Parts: systemParts,
		},
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: humanParts,
		},
		{
			Role:  llms.ChatMessageTypeSystem,
			Parts: systemPartsTwo,
		},
	}

	rsp, err := llm.GenerateContent(context.Background(), content)
	require.Error(t, err)
	require.ErrorIs(t, err, ErrSystemMoreThanOne)
	require.Nil(t, rsp)
}
