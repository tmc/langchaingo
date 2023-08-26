package agents

import (
	"context"
	"fmt"
	"strings"

	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/tools"
)

const _intermediateStepsOutputKey = "intermediateSteps"

// Executor is the chain responsible for running agents.
type Executor struct {
	Agent            Agent
	Tools            []tools.Tool
	Memory           schema.Memory
	CallbacksHandler callbacks.Handler

	MaxIterations           int
	ReturnIntermediateSteps bool
}

var (
	_ chains.Chain           = Executor{}
	_ callbacks.HandlerHaver = Executor{}
)

// NewExecutor creates a new agent executor with a agent and the tools the agent can use.
func NewExecutor(agent Agent, tools []tools.Tool, opts ...CreationOption) Executor {
	options := executorDefaultOptions()
	for _, opt := range opts {
		opt(&options)
	}

	return Executor{
		Agent:                   agent,
		Tools:                   tools,
		Memory:                  options.memory,
		MaxIterations:           options.maxIterations,
		ReturnIntermediateSteps: options.returnIntermediateSteps,
		CallbacksHandler:        options.callbacksHandler,
	}
}

func (e Executor) Call(ctx context.Context, inputValues map[string]any, _ ...chains.ChainCallOption) (map[string]any, error) { //nolint:lll
	inputs, err := inputsToString(inputValues)
	if err != nil {
		return nil, err
	}
	nameToTool := getNameToTool(e.Tools)

	steps := make([]schema.AgentStep, 0)
	for i := 0; i < e.MaxIterations; i++ {
		actions, finish, err := e.Agent.Plan(ctx, steps, inputs)
		if err != nil {
			return nil, err
		}

		if len(actions) == 0 && finish == nil {
			return nil, ErrAgentNoReturn
		}

		if finish != nil {
			return e.getReturn(finish, steps), nil
		}

		for _, action := range actions {
			steps, err = e.doAction(ctx, steps, nameToTool, action)
			if err != nil {
				return nil, err
			}
		}
	}

	return nil, ErrNotFinished
}

func (e Executor) doAction(
	ctx context.Context,
	steps []schema.AgentStep,
	nameToTool map[string]tools.Tool,
	action schema.AgentAction,
) ([]schema.AgentStep, error) {
	if e.CallbacksHandler != nil {
		e.CallbacksHandler.HandleAgentAction(ctx, action)
	}

	tool, ok := nameToTool[strings.ToUpper(action.Tool)]
	if !ok {
		return append(steps, schema.AgentStep{
			Action:      action,
			Observation: fmt.Sprintf("%s is not a valid tool, try another one", action.Tool),
		}), nil
	}

	observation, err := tool.Call(ctx, action.ToolInput)
	if err != nil {
		return nil, err
	}

	return append(steps, schema.AgentStep{
		Action:      action,
		Observation: observation,
	}), nil
}

func (e Executor) getReturn(finish *schema.AgentFinish, steps []schema.AgentStep) map[string]any {
	if e.ReturnIntermediateSteps {
		finish.ReturnValues[_intermediateStepsOutputKey] = steps
	}

	return finish.ReturnValues
}

// GetInputKeys gets the input keys the agent of the executor expects.
// Often "input".
func (e Executor) GetInputKeys() []string {
	return e.Agent.GetInputKeys()
}

// GetOutputKeys gets the output keys the agent of the executor returns.
func (e Executor) GetOutputKeys() []string {
	return e.Agent.GetOutputKeys()
}

func (e Executor) GetMemory() schema.Memory { //nolint:ireturn
	return e.Memory
}

func (e Executor) GetCallbackHandler() callbacks.Handler { //nolint:ireturn
	return e.CallbacksHandler
}

func inputsToString(inputValues map[string]any) (map[string]string, error) {
	inputs := make(map[string]string, len(inputValues))
	for key, value := range inputValues {
		valueStr, ok := value.(string)
		if !ok {
			return nil, fmt.Errorf("%w: %s", ErrExecutorInputNotString, key)
		}

		inputs[key] = valueStr
	}

	return inputs, nil
}

func getNameToTool(t []tools.Tool) map[string]tools.Tool {
	if len(t) == 0 {
		return nil
	}

	nameToTool := make(map[string]tools.Tool, len(t))
	for _, tool := range t {
		nameToTool[strings.ToUpper(tool.Name())] = tool
	}

	return nameToTool
}
