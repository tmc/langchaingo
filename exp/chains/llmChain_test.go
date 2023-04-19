package chains

import (
	"os"
	"strings"
	"testing"

	"github.com/tmc/langchaingo/exp/prompts"
	"github.com/tmc/langchaingo/llms/openai"
)

func TestLLMChain(t *testing.T) {
	if openaiKey := os.Getenv("OPENAI_API_KEY"); openaiKey == "" {
		t.Skip("OPENAI_API_KEY not set")
	}
	model, err := openai.New()
	if err != nil {
		t.Fatal(err)
	}

	prompt, err := prompts.NewPromptTemplate("What is the capital of {country}", []string{"country"})
	if err != nil {
		t.Fatal(err)
	}

	chain := NewLLMChain(model, prompt)

	resultChainValue, err := Call(chain, map[string]any{"country": "France"})
	if err != nil {
		t.Fatal(err)
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
