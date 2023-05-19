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
	// ErrUnableToParseOutput is returned if the output of the llm is unparsable.
	ErrUnableToParseOutput = errors.New("unable to parse agent output")
	// ErrInvalidChainReturnType is returned if the internal chain of the agent
	// returns a value in the "text" filed that is not a string.
	ErrInvalidChainReturnType = errors.New("agent chain did not return a string")
	// ErrInvalidOptions is returned if the options given to the NewOneShotAgent
	// function is invalid.
	ErrInvalidOptions = errors.New("options given are invalid")
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
	// Chain is the chain used to call with the values. The chain should have an
	// input called "agent_scratchpad" for the agent to put it's thoughts in.
	Chain chains.Chain
	// Tools is a list of the tools the agent can use.
	Tools []tools.Tool
	// Output key is the key where the final output is placed.
	OutputKey string
}

var _ agent.Agent = (*OneShotZeroAgent)(nil)

// OneShotZeroAgentOptions is a type alias for a map of string keys to any value,
// representing the options for the OneShotZeroAgent.
type OneShotZeroAgentOptions map[string]any

func checkOptions(opts OneShotZeroAgentOptions) OneShotZeroAgentOptions {
	if _, ok := opts["outputKey"].(string); !ok {
		opts["outputKey"] = _defaultOutputKey
	}
	return opts
}

// NewOneShotAgent creates a new OneShotZeroAgent with the given LLM model, tools,
// and options. It returns a pointer to the created agent.
func NewOneShotAgent(llm llms.LLM, tools []tools.Tool, opts map[string]any) *OneShotZeroAgent {
	// Validate opts.
	opts = checkOptions(opts)

	return &OneShotZeroAgent{
		Chain:     chains.NewLLMChain(llm, CreatePrompt(tools)),
		Tools:     tools,
		OutputKey: opts["outputKey"].(string),
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
		return nil, nil, ErrInvalidChainReturnType
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
				a.OutputKey: splits[len(splits)-1],
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
