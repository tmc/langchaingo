package constitution

import (
	"context"
	"errors"
	"strings"

	"github.com/starmvp/langchaingo/chains"
	"github.com/starmvp/langchaingo/llms"
	"github.com/starmvp/langchaingo/memory"
	"github.com/starmvp/langchaingo/prompts"
	"github.com/starmvp/langchaingo/schema"
)

var (
	// ErrResponseTextNotFound is returned when the response does not provide 'text' key as the result.
	ErrResponseTextNotFound = errors.New("result as a string value of a 'text' key is not found")
	// ErrStringConvert is returned when a value cannot be converted into a string.
	ErrStringConvert = errors.New("cannot convert the provided value to string")
)

type pair struct {
	first, second interface{}
}

// ConstitutionalPrinciple provides critiqueRequest and revisionRequest to be used in a Constitutional instance.
type ConstitutionalPrinciple struct {
	critiqueRequest string
	revisionRequest string
	name            string
}

// Constitutional is a data structure for providing chains of critique and revision to an ordinary chain based on a
// list of ConstitutionalPrinciple.
type Constitutional struct {
	chain                    chains.LLMChain
	critiqueChain            chains.LLMChain
	revisionChain            chains.LLMChain
	constitutionalPrinciples []ConstitutionalPrinciple
	llm                      llms.Model
	returnIntermediateSteps  bool
	memory                   schema.Memory
}

// NewConstitutionalPrinciple creates a new ConstitutionalPrinciple.
func NewConstitutionalPrinciple(critique, revision string, names ...string) ConstitutionalPrinciple {
	var name string
	if len(names) == 0 {
		name = "Constitutional Principle"
	} else {
		name = names[0]
	}
	return ConstitutionalPrinciple{
		critiqueRequest: critique,
		revisionRequest: revision,
		name:            name,
	}
}

// NewConstitutional creates a new Constitutional chain.
func NewConstitutional(llm llms.Model, chain chains.LLMChain,
	constitutionalPrinciples []ConstitutionalPrinciple, options map[string]*prompts.FewShotPrompt,
) *Constitutional {
	CritiquePrompt, RevisionPrompt := initCritiqueRevision()
	var critiquePrompt, revisionPrompt *prompts.FewShotPrompt
	if len(options) == 0 {
		critiquePrompt = CritiquePrompt
		revisionPrompt = RevisionPrompt
	} else {
		var ok bool
		critiquePrompt, ok = options["critique"]
		if !ok {
			critiquePrompt = CritiquePrompt
		}
		revisionPrompt, ok = options["revision"]
		if !ok {
			revisionPrompt = RevisionPrompt
		}
	}

	critiqueChain := *chains.NewLLMChain(llm, critiquePrompt)
	revisionChain := *chains.NewLLMChain(llm, revisionPrompt)

	return &Constitutional{
		chain:                    chain,
		critiqueChain:            critiqueChain,
		revisionChain:            revisionChain,
		constitutionalPrinciples: constitutionalPrinciples,
		llm:                      llm,
		returnIntermediateSteps:  false,
		memory:                   memory.NewSimple(),
	}
}

// Call handles the inner logic of the Constitutional chain.
func (c *Constitutional) Call(ctx context.Context, inputs map[string]any,
	options ...chains.ChainCallOption,
) (map[string]any, error) {
	result, err := c.chain.Call(ctx, inputs, options...)
	if err != nil {
		return nil, err
	}

	response, ok := result["text"]
	if !ok {
		return nil, ErrResponseTextNotFound
	}
	initialResponse := response
	inputPrompt, err := c.chain.Prompt.FormatPrompt(inputs)
	if err != nil {
		return nil, err
	}
	critiquesAndRevisions, err := c.processCritiquesAndRevisions(ctx, response, inputPrompt, options)
	if err != nil {
		return nil, err
	}
	finalOutput := map[string]any{"output": response}
	if c.returnIntermediateSteps {
		finalOutput["initial_output"] = initialResponse
		finalOutput["critiques_and_revisions"] = critiquesAndRevisions
	}
	return finalOutput, nil
}

// processCritiquesAndRevisions processes critiques and revisions based on the input response and prompt.
// It iterates through constitutional principles, retrieves critiques, and performs revisions where necessary.
// The resulting pairs of critiques and revisions are returned.
func (c *Constitutional) processCritiquesAndRevisions(ctx context.Context, response any, inputPrompt llms.PromptValue,
	options []chains.ChainCallOption,
) ([]pair, error) {
	critiquesAndRevisions := make([]pair, 0, len(c.constitutionalPrinciples))
	for _, constitutionalPrincipal := range c.constitutionalPrinciples {
		rawCritique, err := c.critiqueChain.Call(ctx, map[string]any{
			"inputPrompt":     inputPrompt,
			"outputFromModel": response,
			"critiqueRequest": constitutionalPrincipal.critiqueRequest,
		}, options...)
		if err != nil {
			return nil, err
		}
		output, ok := rawCritique["text"]
		if !ok {
			return nil, ErrResponseTextNotFound
		}
		stringOutput, ok := output.(string)
		if !ok {
			return nil, ErrStringConvert
		}
		critique := parseCritique(stringOutput)

		critique = strings.Trim(critique, " ")
		if critique == "no critique needed" {
			continue
		}

		if strings.Contains(strings.ToLower(critique), "no critique needed") {
			critiquesAndRevisions = append(critiquesAndRevisions, pair{
				first:  critique,
				second: "",
			})
			continue
		}

		result, err := c.revisionChain.Call(ctx, map[string]any{
			"inputPrompt":     inputPrompt,
			"outputFromModel": response,
			"critiqueRequest": constitutionalPrincipal.critiqueRequest,
			"critique":        critique,
			"revisionRequest": constitutionalPrincipal.revisionRequest,
		})
		if err != nil {
			return nil, err
		}
		revision, ok := result["text"].(string)
		if !ok {
			return nil, ErrResponseTextNotFound
		}
		revision = strings.Trim(revision, " ")
		response = revision
		critiquesAndRevisions = append(critiquesAndRevisions, pair{
			first:  critique,
			second: revision,
		})
	}
	return critiquesAndRevisions, nil
}

func parseCritique(rawCritique string) string {
	if !strings.Contains(rawCritique, "Revision request:") {
		return rawCritique
	}
	outputString := strings.Split(rawCritique, "Revision request:")[0]
	if strings.Contains(outputString, "\n\n") {
		outputString = strings.Split(outputString, "\n\n")[0]
	}
	return outputString
}

func (c *Constitutional) GetMemory() schema.Memory {
	return c.memory
}

func (c *Constitutional) GetInputKeys() []string {
	return c.chain.GetInputKeys()
}

func (c *Constitutional) GetOutputKeys() []string {
	if c.returnIntermediateSteps {
		return []string{"output", "critiques_and_revisions", "initial_output"}
	}
	return []string{"output"}
}
