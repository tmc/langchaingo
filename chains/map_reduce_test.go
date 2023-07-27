package chains

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/prompts"
	"github.com/tmc/langchaingo/schema"
)

func TestMapReduceInputVariables(t *testing.T) {
	t.Parallel()

	c := MapReduceDocuments{
		LLMChain: NewLLMChain(
			&testLanguageModel{},
			prompts.NewPromptTemplate("{{.text}} {{.foo}}", []string{"text", "foo"}),
		),
		ReduceChain: NewLLMChain(
			&testLanguageModel{},
			prompts.NewPromptTemplate("{{.texts}} {{.baz}}", []string{"texts", "baz"}),
		),
		ReduceDocumentVariableName: "texts",
		LLMChainInputVariableName:  "text",
		InputKey:                   "input",
	}

	inputKeys := c.GetInputKeys()
	expectedLength := 3
	require.Equal(t, expectedLength, len(inputKeys))
}

func TestMapReduce(t *testing.T) {
	t.Parallel()

	c := NewMapReduceDocuments(
		NewLLMChain(
			&testLanguageModel{},
			prompts.NewPromptTemplate("{{.context}}", []string{"context"}),
		),
		NewStuffDocuments(
			NewLLMChain(
				&testLanguageModel{},
				prompts.NewPromptTemplate("{{.context}}", []string{"context"}),
			),
		),
	)

	result, err := Run(context.Background(), c, []schema.Document{
		{PageContent: "foo"},
		{PageContent: "boo"},
		{PageContent: "zoo"},
		{PageContent: "doo"},
	})
	require.NoError(t, err)
	require.Equal(t, "foo\n\nboo\n\nzoo\n\ndoo", result)
}
