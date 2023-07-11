package chains

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/prompts"
	"github.com/tmc/langchaingo/schema"
)

func TestMapRerankInputVariables(t *testing.T) {
	t.Parallel()

	mapRerankLLMChain := NewLLMChain(
		testLanguageModel{},
		prompts.NewPromptTemplate("{{.text}} {{.foo}}", []string{"text", "foo"}),
	)

	c := MapRerankDocuments{
		LLMChain:                  mapRerankLLMChain,
		DocumentVariableName:      "texts",
		LLMChainInputVariableName: "text",
		InputKey:                  "input",
	}

	inputKeys := c.GetInputKeys()
	expectedLength := 3
	require.Equal(t, expectedLength, len(inputKeys))
}

func TestMapRerankDocumentsCall(t *testing.T) {
	t.Parallel()

	mapRerankLLMChain := NewLLMChain(
		testLanguageModel{},
		prompts.NewPromptTemplate("{{.context}}", []string{"context"}),
	)

	docs := []schema.Document{
		{PageContent: "Test Low\nScore: 20"},
		{PageContent: "Test High\nScore: 100"},
	}

	mapRerankDocumentsChain := NewMapRerankDocuments(mapRerankLLMChain)

	// Test that the answer is the highest scoring document.
	answer, err := Run(context.Background(), mapRerankDocumentsChain, docs)

	require.NoError(t, err)
	require.Equal(t, "Test High", answer)

	// Test that the answer cannot be processed if ReturnIntermediateSteps is true.
	mapRerankDocumentsChain.ReturnIntermediateSteps = true
	_, err = Run(context.Background(), mapRerankDocumentsChain, docs)

	require.Error(t, err)
}
