package chains

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
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

func TestConstitutionalChainBasic(t *testing.T) {
	ctx := context.Background()
	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "OPENAI_API_KEY")

	rr := httprr.OpenForTest(t, http.DefaultTransport)

	// Only run tests in parallel when not recording
	if rr.Replaying() {
		t.Parallel()
	}

	opts := []openai.Option{
		openai.WithHTTPClient(rr.Client()),
	}

	// Only add fake token when NOT recording (i.e., during replay)
	if rr.Replaying() {
		opts = append(opts, openai.WithToken("test-api-key"))
	}
	// When recording, openai.New() will read OPENAI_API_KEY from environment

	model, err := openai.New(opts...)
	require.NoError(t, err)
	chain := *NewLLMChain(model, &prompts.FewShotPrompt{
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
	_, err = c.Call(ctx, map[string]any{"question": "What is the meaning of life?"})
	require.NoError(t, err)
}
