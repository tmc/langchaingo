package agents

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/i18n"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/tools"
)

// OneShotZeroAgent is a struct that represents an agent responsible for deciding
// what to do or give the final output if the task is finished given a set of inputs
// and previous steps taken.
//
// This agent is optimized to be used with LLMs.
type OneShotZeroAgent struct {
	// Chain is the chain used to call with the values. The chain should have an
	// input called "agent_scratchpad" for the agent to put its thoughts in.
	Chain chains.Chain
	// Tools is a list of the tools the agent can use.
	Tools []tools.Tool
	// Output key is the key where the final output is placed.
	OutputKey string
	// FinalAnswer is the final answer in various languages.
	FinalAnswer string
	// Lang is the language the prompt will use.
	Lang i18n.Lang
	// CallbacksHandler is the handler for callbacks.
	CallbacksHandler callbacks.Handler
}

var _ Agent = (*OneShotZeroAgent)(nil)

// NewOneShotAgent creates a new OneShotZeroAgent with the given LLM model, tools,
// and options. It returns a pointer to the created agent. The opts parameter
// represents the options for the agent.
func NewOneShotAgent(llm llms.Model, tools []tools.Tool, opts ...Option) *OneShotZeroAgent {
	options := mrklDefaultOptions()
	for _, opt := range opts {
		opt(&options)
	}
	options.loadMrklTranslatable()

	return &OneShotZeroAgent{
		Chain: chains.NewLLMChain(
			llm,
			options.getMrklPrompt(tools),
			chains.WithCallback(options.callbacksHandler),
		),
		Tools:            tools,
		OutputKey:        options.outputKey,
		FinalAnswer:      i18n.AgentsMustPhrase(options.lang, "mrkl final answer"),
		Lang:             options.lang,
		CallbacksHandler: options.callbacksHandler,
	}
}

// Plan decides what action to take or returns the final result of the input.
func (a *OneShotZeroAgent) Plan(
	ctx context.Context,
	intermediateSteps []schema.AgentStep,
	inputs map[string]string,
) ([]schema.AgentAction, *schema.AgentFinish, error) {
	fullInputs := make(map[string]any, len(inputs))
	for key, value := range inputs {
		fullInputs[key] = value
	}

	fullInputs["agent_scratchpad"] = constructMrklScratchPad(intermediateSteps, a.Lang)
	fullInputs["today"] = time.Now().Format(i18n.AgentsMustPhrase(a.Lang, "today format"))

	var stream func(ctx context.Context, chunk []byte) error

	if a.CallbacksHandler != nil {
		stream = func(ctx context.Context, chunk []byte) error {
			a.CallbacksHandler.HandleStreamingFunc(ctx, chunk)
			return nil
		}
	}

	output, err := chains.Predict(
		ctx,
		a.Chain,
		fullInputs,
		chains.WithStopWords([]string{
			fmt.Sprintf("\n%s", i18n.AgentsMustPhrase(a.Lang, "observation")),
			fmt.Sprintf("\n\t%s", i18n.AgentsMustPhrase(a.Lang, "observation")),
		}),
		chains.WithStreamingFunc(stream),
	)
	if err != nil {
		return nil, nil, err
	}

	return a.parseOutput(output)
}

func (a *OneShotZeroAgent) GetInputKeys() []string {
	chainInputs := a.Chain.GetInputKeys()

	// Remove inputs given in plan.
	agentInput := make([]string, 0, len(chainInputs))
	for _, v := range chainInputs {
		if v == "agent_scratchpad" || v == "today" {
			continue
		}
		agentInput = append(agentInput, v)
	}

	return agentInput
}

func (a *OneShotZeroAgent) GetOutputKeys() []string {
	return []string{a.OutputKey}
}

func (a *OneShotZeroAgent) GetTools() []tools.Tool {
	return a.Tools
}

func constructMrklScratchPad(steps []schema.AgentStep, lang i18n.Lang) string {
	var scratchPad string
	if len(steps) > 0 {
		for _, step := range steps {
			scratchPad += "\n" + step.Action.Log
			scratchPad += fmt.Sprintf("\n%s %s\n", i18n.AgentsMustPhrase(lang, "observation"), step.Observation)
		}
	}

	return scratchPad
}

func (a *OneShotZeroAgent) parseOutput(output string) ([]schema.AgentAction, *schema.AgentFinish, error) {
	if strings.Contains(output, a.FinalAnswer) {
		splits := strings.Split(output, a.FinalAnswer)

		return nil, &schema.AgentFinish{
			ReturnValues: map[string]any{
				a.OutputKey: splits[len(splits)-1],
			},
			Log: output,
		}, nil
	}

	action, actionInput, observation := i18n.AgentsMustPhrase(a.Lang, "action"),
		i18n.AgentsMustPhrase(a.Lang, "action input"),
		i18n.AgentsMustPhrase(a.Lang, "observation")
	var r *regexp.Regexp
	if strings.Contains(output, observation) {
		r = regexp.MustCompile(fmt.Sprintf(`%s\s*(.+)\s*%s\s(?s)*(.+)%s`, action, actionInput, observation))
	} else {
		r = regexp.MustCompile(fmt.Sprintf(`%s\s*(.+)\s*%s\s(?s)*(.+)`, action, actionInput))
	}
	matches := r.FindStringSubmatch(output)
	if len(matches) == 0 {
		return nil, nil, fmt.Errorf("%w: %s", ErrUnableToParseOutput, output)
	}

	return []schema.AgentAction{
		{Tool: strings.TrimSpace(matches[1]), ToolInput: strings.TrimSpace(matches[2]), Log: output},
	}, nil, nil
}
