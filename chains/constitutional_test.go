package chains

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/internal/httprr"
	"github.com/tmc/langchaingo/llms/fake"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/prompts"
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

func TestConstitutionalRevisedOutput(t *testing.T) {
	t.Parallel()
	p := NewConstitutionalPrinciple("c", "r")
	prompt := prompts.NewPromptTemplate("{{.q}}", []string{"q"})
	tests := []struct {
		name, want string
		prins      []ConstitutionalPrinciple
		replies    []string
		pairs      [][2]string
	}{
		{"no principles/no revision", "initial", nil, []string{"initial"}, nil},
		{"single revision", "rev1", []ConstitutionalPrinciple{p},
			[]string{"initial", "needs fix", "rev1"}, [][2]string{{"needs fix", "rev1"}}},
		{"multiple revisions", "rev2", []ConstitutionalPrinciple{p, p},
			[]string{"initial", "bad", "rev1", "worse", "rev2"},
			[][2]string{{"bad", "rev1"}, {"worse", "rev2"}}},
		{"revision then no-critique", "rev1", []ConstitutionalPrinciple{p, p},
			[]string{"initial", "bad", "rev1", "No critique needed."},
			[][2]string{{"bad", "rev1"}, {"No critique needed.", ""}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			llm := fake.NewFakeLLM(tt.replies)
			c := NewConstitutional(llm, *NewLLMChain(llm, prompt), tt.prins, nil)
			c.returnIntermediateSteps = true
			got, err := c.Call(context.Background(), map[string]any{"q": "hi"})
			require.NoError(t, err)
			require.Equal(t, "initial", got["initial_output"])
			require.Equal(t, tt.want, got["output"])
			raw, _ := got["critiques_and_revisions"].([]Pair)
			require.Len(t, raw, len(tt.pairs))
			for i, pair := range tt.pairs {
				require.Equal(t, pair[0], raw[i].first)
				require.Equal(t, pair[1], raw[i].second)
			}
		})
	}
}
