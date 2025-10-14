package chains

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vendasta/langchaingo/internal/httprr"
	"github.com/vendasta/langchaingo/llms/openai"
	"github.com/vendasta/langchaingo/prompts"
	"github.com/vendasta/langchaingo/schema"
)

func TestStuffDocuments(t *testing.T) {
	ctx := context.Background()

	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "OPENAI_API_KEY")

	rr := httprr.OpenForTest(t, http.DefaultTransport)

	// Only run tests in parallel when not recording
	if rr.Replaying() {
		t.Parallel()
	}

	opts := []openai.Option{
		openai.WithHTTPClient(rr.Client()),
	}

	// Only add fake token when NOT recording (i.e., during replay)
	if rr.Replaying() {
		opts = append(opts, openai.WithToken("test-api-key"))
	}
	// When recording, openai.New() will read OPENAI_API_KEY from environment

	model, err := openai.New(opts...)
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
