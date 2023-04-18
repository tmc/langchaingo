package chains

import (
	"strings"
	"testing"

	"github.com/tmc/langchaingo/exp/prompts"
	"github.com/tmc/langchaingo/llms/openai"
)

func TestLLMChain(t *testing.T) {
	model, err := openai.New()
	if err != nil {
		t.Errorf("Unexpected error creating openAI model: %e", err)
		return
	}

	prompt, err := prompts.NewPromptTemplate("What is the capital of {country}", []string{"country"})
	if err != nil {
		t.Errorf("Unexpected error creating prompt template: %e", err)
		return
	}

	chain := NewLLMChain(model, prompt)

	resultChainValue, err := Call(chain, map[string]any{"country": "France"})
	if err != nil {
		t.Errorf("Unexpected error calling llm chain: %e", err)
		return
	}

	resultAny, ok := resultChainValue["text"]
	if !ok {
		t.Error("No value in chain value text field")
		return
	}

	result, ok := resultAny.(string)
	result = strings.TrimSpace(result)

	if result != "Paris." {
		t.Errorf("Expected to get Paris. Got %s", result)
	}
}
