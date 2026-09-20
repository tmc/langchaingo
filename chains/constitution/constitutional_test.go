package constitution

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/internal/httprr"
	"github.com/tmc/langchaingo/llms/fake"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/prompts"
)

// hasExistingRecording checks if a httprr recording exists for this test
func hasExistingRecording(t *testing.T) bool {
	testName := strings.ReplaceAll(t.Name(), "/", "_")
	testName = strings.ReplaceAll(testName, " ", "_")
	recordingPath := filepath.Join("testdata", testName+".httprr")
	_, err := os.Stat(recordingPath)
	return err == nil
}

func TestConstitutionCritiqueParsing(t *testing.T) {

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

func TestConstitutionalChain(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// Skip if no recording available and no credentials
	if !hasExistingRecording(t) {
		t.Skip("No httprr recording available. Hint: Re-run tests with -httprecord=. to record new HTTP interactions")
	}

	rr := httprr.OpenForTest(t, http.DefaultTransport)
	opts := []openai.Option{
		openai.WithHTTPClient(rr.Client()),
	}
	if rr.Replaying() {
		opts = append(opts, openai.WithToken("test-api-key"))
	}

	model, err := openai.New(opts...)
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
	_, err = c.Call(ctx, map[string]any{"question": "What is the meaning of life?"})
	if err != nil {
		// Check if this is a recording mismatch error
		if strings.Contains(err.Error(), "cached HTTP response not found") {
			t.Skip("Recording format has changed or is incompatible. Hint: Re-run tests with -httprecord=. to record new HTTP interactions")
		}
		require.NoError(t, err)
	}
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
			c := NewConstitutional(llm, *chains.NewLLMChain(llm, prompt), tt.prins, nil)
			c.returnIntermediateSteps = true
			got, err := c.Call(context.Background(), map[string]any{"q": "hi"})
			require.NoError(t, err)
			require.Equal(t, "initial", got["initial_output"])
			require.Equal(t, tt.want, got["output"])
			raw, _ := got["critiques_and_revisions"].([]pair)
			require.Len(t, raw, len(tt.pairs))
			for i, pair := range tt.pairs {
				require.Equal(t, pair[0], raw[i].first)
				require.Equal(t, pair[1], raw[i].second)
			}
		})
	}
}
