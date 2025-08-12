package constitution

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vendasta/langchaingo/chains"
	"github.com/vendasta/langchaingo/llms/openai"
	"github.com/vendasta/langchaingo/prompts"
)

func TestConstitutionCritiqueParsing(t *testing.T) {
	t.Parallel()
	textOne := ` This text is bad.

	Revision request: Make it better.
	
	Revision:`

	textTwo := " This text is bad.\n\n"

	textThree := ` This text is bad.
	
	Revision request: Make it better.
	
	Revision: Better text`

	for _, rawCritique := range []string{textOne, textTwo, textThree} {
		critique := parseCritique(rawCritique)
		require.Equal(t, "This text is bad.", strings.TrimSpace(critique),
			fmt.Sprintf("Failed on %s with %s", rawCritique, critique))
	}
}

func Test(t *testing.T) {
	t.Parallel()
	if openaiKey := os.Getenv("OPENAI_API_KEY"); openaiKey == "" {
		t.Skip("OPENAI_API_KEY not set")
	}
	model, err := openai.New()
	require.NoError(t, err)
	chain := *chains.NewLLMChain(model, &prompts.FewShotPrompt{
		Examples:         []map[string]string{{"question": "What's life?"}},
		ExampleSelector:  nil,
		ExamplePrompt:    prompts.NewPromptTemplate("{{.question}}", []string{"question"}),
		Prefix:           "",
		Suffix:           "",
		InputVariables:   []string{"question"},
		PartialVariables: nil,
		TemplateFormat:   prompts.TemplateFormatGoTemplate,
		ValidateTemplate: false,
	})

	c := NewConstitutional(model, chain, []ConstitutionalPrinciple{
		NewConstitutionalPrinciple(
			"Tell if this answer is good.",
			"Give a better answer.",
		),
	}, nil)
	_, err = c.Call(context.Background(), map[string]any{"question": "What is the meaning of life?"})
	require.NoError(t, err)
}
