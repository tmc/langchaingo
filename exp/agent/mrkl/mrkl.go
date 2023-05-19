package mrkl

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/exp/agent"
	"github.com/tmc/langchaingo/exp/tools"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/schema"
)

var (
	ErrUnableToParseOutput = errors.New("unable to parse agent output")
)

const (
	_finalAnswerAction = "Final Answer:"
	_defaultOutputKey  = "output"
)

// OneShotZeroAgent is a struct that represents an agent responsible for deciding
// what to do or give the final output if the task is finished given a set of inputs
// and previous steps taken.
//
// This agent is optimized to be used with LLMs.
type OneShotZeroAgent struct {
	// Chain is the chain used to call with the values
	Chain chains.Chain
	// Tools is the slice of the tools the agent can use.
	Tools []tools.Tool

	maxRetries int
	outputKey  string
}

var _ agent.Agent = (*OneShotZeroAgent)(nil)

// OneShotZeroAgentOptions is a type alias for a map of string keys to any value,
// representing the options for the OneShotZeroAgent.
type OneShotZeroAgentOptions map[string]any

func checkOptions(opts OneShotZeroAgentOptions) OneShotZeroAgentOptions {
	if _, ok := opts["maxRetries"].(int); !ok {
		opts["maxRetries"] = 3
	}
	return opts
}

// NewOneShotAgent creates a new OneShotZeroAgent with the given LLM model, tools,
// and options. It returns a pointer to the created agent.
func NewOneShotAgent(llm llms.LLM, tools []tools.Tool, opts map[string]any) *OneShotZeroAgent {
	// Validate opts
	opts = checkOptions(opts)

	return &OneShotZeroAgent{
		Chain:      chains.NewLLMChain(llm, createPrompt(tools)),
		Tools:      tools,
		maxRetries: opts["maxRetries"].(int),
		outputKey:  _defaultOutputKey,
	}
}

// Plan decides what to do or returns the result of the input.
func (a *OneShotZeroAgent) Plan(
	ctx context.Context,
	intermediateSteps []schema.AgentStep,
	inputs map[string]string,
) ([]schema.AgentAction, *schema.AgentFinish, error) {
	fullInputs := make(map[string]any, len(inputs))
	for key, value := range inputs {
		fullInputs[key] = value
	}

	fullInputs["agent_scratchpad"] = a.constructScratchPad(intermediateSteps)
	fullInputs["today"] = time.Now().Format("January 02, 2006")

	resp, err := chains.Call(
		ctx,
		a.Chain,
		fullInputs,
		chains.WithStopWords([]string{"\nObservation:", "\n\tObservation:"}),
	)
	if err != nil {
		return nil, nil, err
	}

	output, ok := resp["text"].(string)
	if !ok {
		return nil, nil, errors.New("Agent did not return a string")
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
	return []string{a.outputKey}
}

func (a *OneShotZeroAgent) constructScratchPad(steps []schema.AgentStep) string {
	var scratchPad string
	for _, step := range steps {
		scratchPad += step.Action.Log
		scratchPad += "Observation: " + step.Observation
	}

	return scratchPad
}

func (a *OneShotZeroAgent) parseOutput(output string) ([]schema.AgentAction, *schema.AgentFinish, error) {
	if strings.Contains(output, _finalAnswerAction) {
		splits := strings.Split(output, _finalAnswerAction)

		return nil, &schema.AgentFinish{
			ReturnValues: map[string]any{
				"output": splits[len(splits)-1],
			},
			Log: output,
		}, nil
	}

	r := regexp.MustCompile(`Action:\s*(.+)\s*Action Input:\s*(.+)`)
	matches := r.FindStringSubmatch(output)
	if len(matches) == 0 {
		return nil, nil, fmt.Errorf("%w: %s", ErrUnableToParseOutput, output)
	}

	return []schema.AgentAction{
		{Tool: strings.TrimSpace(matches[1]), ToolInput: strings.TrimSpace(matches[2]), Log: output},
	}, nil, nil
}
