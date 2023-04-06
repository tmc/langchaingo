package chains_test

import (
	"testing"

	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/prompts"
)

func TestLLMChain(t *testing.T) {
	model, err := openai.New()
	if err != nil {
		t.Errorf("Unexpected error creating openAI model: %e", err)
	}

	prompt, err := prompts.NewPromptTemplate("What is the capital of {country}", []string{"country"})
	if err != nil {
		t.Errorf("Unexpected error creating prompt template: %e", err)
	}

	chain := chains.NewLLMChain(model, prompt)

	resultChainValue, err := chains.Call(chain, map[string]any{"country": "France"})
	if err != nil {
		t.Errorf("Unexpected error calling llm chain: %e", err)
	}

	result, ok := resultChainValue["text"]
	if !ok {
		t.Error("No value in chain value text field")
	}

	t.Logf("Expected result Paris. Result gotten %s", result)
}
