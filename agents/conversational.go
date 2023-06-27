package agents

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/tools"
)

const (
	_conversationalFinalAnswerAction = "AI:"
)

// ConversationalAgent is a struct that represents an agent responsible for deciding
// what to do or give the final output if the task is finished given a set of inputs
// and previous steps taken.
//
// Other agents are often optimized for using tools to figure out the best response,
// which is not ideal in a conversational setting where you may want the agent to be
// able to chat with the user as well.
type ConversationalAgent struct {
	// Chain is the chain used to call with the values. The chain should have an
	// input called "agent_scratchpad" for the agent to put it's thoughts in.
	Chain chains.Chain
	// Tools is a list of the tools the agent can use.
	Tools []tools.Tool
	// Output key is the key where the final output is placed.
	OutputKey string
}

var _ Agent = (*ConversationalAgent)(nil)

func NewConversationalAgent(llm llms.LanguageModel, tools []tools.Tool, opts ...CreationOption) *ConversationalAgent {
	options := conversationalDefaultOptions()
	for _, opt := range opts {
		opt(&options)
	}

	return &ConversationalAgent{
		Chain:     chains.NewLLMChain(llm, options.getConversationalPrompt(tools)),
		Tools:     tools,
		OutputKey: options.outputKey,
	}
}

// Plan decides what action to take or returns the final result of the input.
func (a *ConversationalAgent) Plan(
	ctx context.Context,
	intermediateSteps []schema.AgentStep,
	inputs map[string]string,
) ([]schema.AgentAction, *schema.AgentFinish, error) {
	fullInputs := make(map[string]any, len(inputs))
	for key, value := range inputs {
		fullInputs[key] = value
	}

	fullInputs["agent_scratchpad"] = constructScratchPad(intermediateSteps)

	output, err := chains.Predict(
		ctx,
		a.Chain,
		fullInputs,
		chains.WithStopWords([]string{"\nObservation:", "\n\tObservation:"}),
	)
	if err != nil {
		return nil, nil, err
	}

	return a.parseOutput(output)
}

func (a *ConversationalAgent) GetInputKeys() []string {
	chainInputs := a.Chain.GetInputKeys()

	// Remove inputs given in plan.
	agentInput := make([]string, 0, len(chainInputs))
	for _, v := range chainInputs {
		if v == "agent_scratchpad" {
			continue
		}
		agentInput = append(agentInput, v)
	}

	return agentInput
}

func (a *ConversationalAgent) GetOutputKeys() []string {
	return []string{a.OutputKey}
}

func constructScratchPad(steps []schema.AgentStep) string {
	var scratchPad string
	if len(steps) > 0 {
		for _, step := range steps {
			scratchPad += step.Action.Log
			scratchPad += "\nObservation: " + step.Observation
		}
		scratchPad += "\n" + "Thought:"
	}

	return scratchPad
}

func (a *ConversationalAgent) parseOutput(output string) ([]schema.AgentAction, *schema.AgentFinish, error) {
	if strings.Contains(output, _conversationalFinalAnswerAction) {
		splits := strings.Split(output, _conversationalFinalAnswerAction)

		return nil, &schema.AgentFinish{
			ReturnValues: map[string]any{
				a.OutputKey: splits[len(splits)-1],
			},
			Log: output,
		}, nil
	}

	r := regexp.MustCompile(`Action: (.*?)[\n]*Action Input: (.*)`)
	matches := r.FindStringSubmatch(output)
	if len(matches) == 0 {
		return nil, nil, fmt.Errorf("%w: %s", ErrUnableToParseOutput, output)
	}

	return []schema.AgentAction{
		{Tool: strings.TrimSpace(matches[1]), ToolInput: strings.TrimSpace(matches[2]), Log: output},
	}, nil, nil
}
