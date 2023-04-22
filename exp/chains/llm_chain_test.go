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

	resultChainValue, err := Call(chain, map[string]any{"country": "France", "stop": []string{"\nObservation", "\n\tObservation"}})
	if err != nil {
		t.Fatal(err)
	}

	resultAny, ok := resultChainValue["text"]
	if !ok {
		t.Fatal("No value in chain value text field")
		return
	}

	result, _ := resultAny.(string)
	result = strings.TrimSpace(result)

	if result != "Paris." {
		t.Fatalf("Expected to get Paris. Got %s", result)
	}
}
