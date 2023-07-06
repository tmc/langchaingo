package chains

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/prompts"
)

func TestMapReduceInputVariables(t *testing.T) {
	t.Parallel()

	c := MapReduceDocuments{
		LLMChain: NewLLMChain(
			testLanguageModel{},
			prompts.NewPromptTemplate("{{.text}} {{.foo}}", []string{"text", "foo"}),
		),
		ReduceChain: NewLLMChain(
			testLanguageModel{},
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
