package chains

import (
	"testing"

	"github.com/tmc/langchaingo/exp/prompts"
	"github.com/tmc/langchaingo/exp/schema"
	"github.com/tmc/langchaingo/llms/openai"
)

func TestStuffDocumentsChain(t *testing.T) {
	model, err := openai.New()
	if err != nil {
		t.Errorf("Unexpected error %s", err)
		return
	}

	prompt, err := prompts.NewPromptTemplate(
		"Print {context}",
		[]string{"context"},
	)
	if err != nil {
		t.Errorf("Unexpected error %s", err)
		return
	}

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
	if err != nil {
		t.Errorf("Unexpected error %s", err.Error())
		return
	}
}
