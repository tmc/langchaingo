package chains

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/starmvp/langchaingo/llms/openai"
	"github.com/starmvp/langchaingo/prompts"
	"github.com/starmvp/langchaingo/schema"
)

func TestStuffDocuments(t *testing.T) {
	t.Parallel()

	if openaiKey := os.Getenv("OPENAI_API_KEY"); openaiKey == "" {
		t.Skip("OPENAI_API_KEY not set")
	}
	model, err := openai.New()
	require.NoError(t, err)

	prompt := prompts.NewPromptTemplate(
		"Write {{.context}}",
		[]string{"context"},
	)
	require.NoError(t, err)

	llmChain := NewLLMChain(model, prompt)
	chain := NewStuffDocuments(llmChain)

	docs := []schema.Document{
		{PageContent: "foo"},
		{PageContent: "bar"},
		{PageContent: "baz"},
	}

	result, err := Call(context.Background(), chain, map[string]any{
		"input_documents": docs,
	})
	require.NoError(t, err)
	for _, key := range chain.GetOutputKeys() {
		_, ok := result[key]
		require.True(t, ok)
	}
}

func TestStuffDocuments_joinDocs(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		name string
		docs []schema.Document
		want string
	}{
		{
			name: "empty",
			docs: []schema.Document{},
			want: "",
		},
		{
			name: "single",
			docs: []schema.Document{
				{PageContent: "foo"},
			},
			want: "foo",
		},
		{
			name: "multiple",
			docs: []schema.Document{
				{PageContent: "foo"},
				{PageContent: "bar"},
			},
			want: "foo\n\nbar",
		},
		{
			name: "multiple with separator",
			docs: []schema.Document{
				{PageContent: "foo"},
				{PageContent: "bar\n\n"},
			},
			want: "foo\n\nbar\n\n",
		},
	}

	chain := NewStuffDocuments(&LLMChain{})

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := chain.joinDocuments(tc.docs)
			require.Equal(t, tc.want, got)
		})
	}
}
