package chains

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vendasta/langchaingo/callbacks"
	"github.com/vendasta/langchaingo/llms/googleai"
	"github.com/vendasta/langchaingo/llms/openai"
	"github.com/vendasta/langchaingo/prompts"
)

func TestLLMChain(t *testing.T) {
	t.Parallel()
	if openaiKey := os.Getenv("OPENAI_API_KEY"); openaiKey == "" {
		t.Skip("OPENAI_API_KEY not set")
	}
	model, err := openai.New()
	require.NoError(t, err)
	model.CallbacksHandler = callbacks.LogHandler{}

	prompt := prompts.NewPromptTemplate(
		"What is the capital of {{.country}}",
		[]string{"country"},
	)
	require.NoError(t, err)

	chain := NewLLMChain(model, prompt)

	result, err := Predict(context.Background(), chain,
		map[string]any{
			"country": "France",
		},
	)
	require.NoError(t, err)
	require.True(t, strings.Contains(result, "Paris"))
}

func TestLLMChainWithChatPromptTemplate(t *testing.T) {
	t.Parallel()

	c := NewLLMChain(
		&testLanguageModel{},
		prompts.NewChatPromptTemplate([]prompts.MessageFormatter{
			prompts.NewAIMessagePromptTemplate("{{.foo}}", []string{"foo"}),
			prompts.NewHumanMessagePromptTemplate("{{.boo}}", []string{"boo"}),
		}),
	)
	result, err := Predict(context.Background(), c, map[string]any{
		"foo": "foo",
		"boo": "boo",
	})
	require.NoError(t, err)
	require.Equal(t, "AI: foo\nHuman: boo", result)
}

func TestLLMChainWithGoogleAI(t *testing.T) {
	t.Parallel()
	genaiKey := os.Getenv("GENAI_API_KEY")
	if genaiKey == "" {
		t.Skip("GENAI_API_KEY not set")
	}
	model, err := googleai.New(context.Background(), googleai.WithAPIKey(genaiKey))
	require.NoError(t, err)
	require.NoError(t, err)
	model.CallbacksHandler = callbacks.LogHandler{}

	prompt := prompts.NewPromptTemplate(
		"What is the capital of {{.country}}",
		[]string{"country"},
	)
	require.NoError(t, err)

	chain := NewLLMChain(model, prompt)

	// chains tramples over defaults for options, so setting these options
	// explicitly is required until https://github.com/vendasta/langchaingo/issues/626
	// is fully resolved.
	result, err := Predict(context.Background(), chain,
		map[string]any{
			"country": "France",
		},
		WithCallback(callbacks.LogHandler{}),
	)
	require.NoError(t, err)
	require.True(t, strings.Contains(result, "Paris"))
}
