package agents

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/tools"
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
	// input called "agent_scratchpad" for the agent to put its thoughts in.
	Chain chains.Chain
	// Tools is a list of the tools the agent can use.
	Tools []tools.Tool
	// Output key is the key where the final output is placed.
	OutputKey string
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

	return &OneShotZeroAgent{
		Chain: chains.NewLLMChain(
			llm,
			options.getMrklPrompt(tools),
			chains.WithCallback(options.callbacksHandler),
		),
		Tools:            tools,
		OutputKey:        options.outputKey,
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

	fullInputs["agent_scratchpad"] = constructMrklScratchPad(intermediateSteps)

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
		chains.WithStopWords([]string{"\nObservation:", "\n\tObservation:"}),
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
		if v == "agent_scratchpad" {
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

func constructMrklScratchPad(steps []schema.AgentStep) string {
	var scratchPad string
	if len(steps) > 0 {
		for _, step := range steps {
			scratchPad += "\n" + step.Action.Log
			scratchPad += "\nObservation: " + step.Observation + "\n"
		}
	}

	return scratchPad
}

func (a *OneShotZeroAgent) parseOutput(output string) ([]schema.AgentAction, *schema.AgentFinish, error) {
	// First check for the standard "Final Answer:" format for backward compatibility
	if strings.Contains(output, _finalAnswerAction) {
		splits := strings.Split(output, _finalAnswerAction)

		return nil, &schema.AgentFinish{
			ReturnValues: map[string]any{
				a.OutputKey: strings.TrimSpace(splits[len(splits)-1]),
			},
			Log: output,
		}, nil
	}

	// Check for case-insensitive variations of final answer
	// This helps with models that may use different casing
	lowerOutput := strings.ToLower(output)
	finalAnswerVariations := []string{
		"final answer:",
		"final answer :",
		"the final answer is:",
		"the answer is:",
	}

	for _, variation := range finalAnswerVariations {
		if idx := strings.Index(lowerOutput, variation); idx != -1 {
			// Extract the answer after the variation phrase
			answerStart := idx + len(variation)
			answer := strings.TrimSpace(output[answerStart:])

			// Make sure this isn't followed by an Action (which would indicate it's not really done)
			if !strings.Contains(strings.ToLower(answer), "\naction:") {
				return nil, &schema.AgentFinish{
					ReturnValues: map[string]any{
						a.OutputKey: answer,
					},
					Log: output,
				}, nil
			}
		}
	}

	// Try to match Action/Action Input pattern with more flexibility
	// Support both "Action Input:" and "Action input:" (case variations)
	r := regexp.MustCompile(`(?i)Action:\s*(.+?)\s*Action\s+Input:\s*(?s)(.+)`)
	matches := r.FindStringSubmatch(output)
	if len(matches) == 3 {
		return []schema.AgentAction{
			{Tool: strings.TrimSpace(matches[1]), ToolInput: strings.TrimSpace(matches[2]), Log: output},
		}, nil, nil
	}

	// Fallback to original regex for backward compatibility
	r = regexp.MustCompile(`Action:\s*(.+)\s*Action Input:\s(?s)*(.+)`)
	matches = r.FindStringSubmatch(output)
	if len(matches) == 0 {
		return nil, nil, fmt.Errorf("%w: %s", ErrUnableToParseOutput, output)
	}

	return []schema.AgentAction{
		{Tool: strings.TrimSpace(matches[1]), ToolInput: strings.TrimSpace(matches[2]), Log: output},
	}, nil, nil
}
