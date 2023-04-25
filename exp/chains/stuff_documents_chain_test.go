package chains

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/prompts"
	"github.com/tmc/langchaingo/schema"
)

func TestStuffDocumentsChain(t *testing.T) {
	t.Parallel()

	if openaiKey := os.Getenv("OPENAI_API_KEY"); openaiKey == "" {
		t.Skip("OPENAI_API_KEY not set")
	}
	model, err := openai.New()
	require.NoError(t, err)

	prompt, err := prompts.NewPromptTemplate(
		"Print {context}",
		[]string{"context"},
	)
	require.NoError(t, err)

	llmChain := NewLLMChain(model, prompt)
	chain := NewStuffDocumentsChain(llmChain)

	docs := []schema.Document{
		{PageContent: "foo", Metadata: map[string]any{}},
		{PageContent: "bar", Metadata: map[string]any{}},
		{PageContent: "baz", Metadata: map[string]any{}},
	}
	inputValues := map[string]any{
		"input_documents": docs,
	}

	_, err = Call(chain, inputValues)
	require.NoError(t, err)
}
