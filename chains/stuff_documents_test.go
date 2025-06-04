package chains

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/internal/httprr"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/prompts"
	"github.com/tmc/langchaingo/schema"
)

func TestStuffDocuments(t *testing.T) {
	ctx := context.Background()
	t.Parallel()

	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "OPENAI_API_KEY")

	rr := httprr.OpenForTest(t, http.DefaultTransport)
	t.Cleanup(func() { rr.Close() })
	model, err := openai.New(openai.WithHTTPClient(rr.Client()))
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

	result, err := Call(ctx, chain, map[string]any{
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
