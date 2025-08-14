package chains

import (
	"context"
	"testing"

	"github.com/0xDezzy/langchaingo/prompts"
	"github.com/0xDezzy/langchaingo/schema"
	"github.com/stretchr/testify/require"
)

func TestMapRerankInputVariables(t *testing.T) {
	t.Parallel()

	mapRerankLLMChain := NewLLMChain(
		&testLanguageModel{},
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
	require.Len(t, inputKeys, expectedLength)
}

func TestMapRerankDocumentsCall(t *testing.T) {
	ctx := context.Background()
	t.Parallel()

	mapRerankLLMChain := NewLLMChain(
		&testLanguageModel{},
		prompts.NewPromptTemplate("{{.context}}", []string{"context"}),
	)

	docs := []schema.Document{
		{PageContent: "Test Low\nScore: 20"},
		{PageContent: "Test High\nScore: 100"},
	}

	mapRerankDocumentsChain := NewMapRerankDocuments(mapRerankLLMChain)

	// Test that the answer is the highest scoring document.
	answer, err := Run(ctx, mapRerankDocumentsChain, docs)

	require.NoError(t, err)
	require.Equal(t, "Test High", answer)

	// Test that the answer cannot be processed if ReturnIntermediateSteps is true.
	mapRerankDocumentsChain.ReturnIntermediateSteps = true
	_, err = Run(ctx, mapRerankDocumentsChain, docs)

	require.Error(t, err)

	// Test that an error is returned if the score cannot be processed.
	mapRerankDocumentsChain.ReturnIntermediateSteps = false
	docs = []schema.Document{
		{PageContent: "Test Low\nScore:"},
		{PageContent: "Test High\nScore:"},
	}

	_, err = Run(ctx, mapRerankDocumentsChain, docs)

	require.Error(t, err)
}
