package chains

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/0xDezzy/langchaingo/callbacks"
	"github.com/0xDezzy/langchaingo/httputil"
	"github.com/0xDezzy/langchaingo/internal/httprr"
	"github.com/0xDezzy/langchaingo/llms/googleai"
	"github.com/0xDezzy/langchaingo/llms/openai"
	"github.com/0xDezzy/langchaingo/prompts"
)

func TestLLMChain(t *testing.T) {
	ctx := context.Background()
	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "OPENAI_API_KEY")

	rr := httprr.OpenForTest(t, httputil.DefaultTransport)

	// Only run tests in parallel when not recording (to avoid rate limits)
	if rr.Replaying() {
		t.Parallel()
	}

	var opts []openai.Option
	opts = append(opts, openai.WithHTTPClient(rr.Client()))

	// Use test token when replaying
	if rr.Replaying() {
		opts = append(opts, openai.WithToken("test-api-key"))
	}

	model, err := openai.New(opts...)
	require.NoError(t, err)
	model.CallbacksHandler = callbacks.LogHandler{}

	prompt := prompts.NewPromptTemplate(
		"What is the capital of {{.country}}",
		[]string{"country"},
	)

	chain := NewLLMChain(model, prompt)

	result, err := Predict(ctx, chain,
		map[string]any{
			"country": "France",
		},
	)
	require.NoError(t, err)
	require.True(t, strings.Contains(result, "Paris"))
}

func TestLLMChainWithChatPromptTemplate(t *testing.T) {
	ctx := context.Background()
	t.Parallel()

	c := NewLLMChain(
		&testLanguageModel{},
		prompts.NewChatPromptTemplate([]prompts.MessageFormatter{
			prompts.NewAIMessagePromptTemplate("{{.foo}}", []string{"foo"}),
			prompts.NewHumanMessagePromptTemplate("{{.boo}}", []string{"boo"}),
		}),
	)
	result, err := Predict(ctx, c, map[string]any{
		"foo": "foo",
		"boo": "boo",
	})
	require.NoError(t, err)
	require.Equal(t, "AI: foo\nHuman: boo", result)
}

func TestLLMChainWithGoogleAI(t *testing.T) {
	ctx := context.Background()
	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "GOOGLE_API_KEY")

	rr := httprr.OpenForTest(t, httputil.DefaultTransport)

	// Configure client with httprr - use test credentials when replaying
	var opts []googleai.Option
	opts = append(opts, googleai.WithRest(), googleai.WithHTTPClient(rr.Client()))

	if rr.Replaying() {
		// Use test credentials during replay
		opts = append(opts, googleai.WithAPIKey("test-api-key"))
	}

	model, err := googleai.New(ctx, opts...)
	require.NoError(t, err)
	model.CallbacksHandler = callbacks.LogHandler{}

	prompt := prompts.NewPromptTemplate(
		"What is the capital of {{.country}}",
		[]string{"country"},
	)

	chain := NewLLMChain(model, prompt)

	// chains tramples over defaults for options, so setting these options
	// explicitly is required until https://github.com/0xDezzy/langchaingo/issues/626
	// is fully resolved.
	result, err := Predict(ctx, chain,
		map[string]any{
			"country": "France",
		},
	)
	require.NoError(t, err)
	require.True(t, strings.Contains(result, "Paris"))
}
