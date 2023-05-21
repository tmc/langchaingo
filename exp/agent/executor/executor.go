package executor

import (
	"context"
	"fmt"

	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/exp/agent"
	"github.com/tmc/langchaingo/exp/tools"
	"github.com/tmc/langchaingo/memory"
	"github.com/tmc/langchaingo/schema"
)

// Executor is the chain responsible for running agents.
type Executor struct {
	Agent agent.Agent
	Tools []tools.Tool

	MaxIterations int
}

var _ chains.Chain = Executor{}

// New creates a new agent executor with a agent, the tools the agent can use
// and the max number of iterations.
func New(agent agent.Agent, tools []tools.Tool, maxIterations int) Executor {
	return Executor{
		Agent:         agent,
		Tools:         tools,
		MaxIterations: maxIterations,
	}
}

func (e Executor) Call(ctx context.Context, inputValues map[string]any, _ ...chains.ChainCallOption) (map[string]any, error) { //nolint:lll
	inputs, err := inputsToString(inputValues)
	if err != nil {
		return nil, err
	}

	nameToTool := make(map[string]tools.Tool, len(e.Tools))
	for _, tool := range e.Tools {
		nameToTool[tool.Name()] = tool
	}

	steps := make([]schema.AgentStep, 0)
	iterations := 0
	for iterations < e.MaxIterations {
		actions, finish, err := e.Agent.Plan(ctx, steps, inputs)
		if err != nil {
			return nil, err
		}

		if len(actions) == 0 && finish == nil {
			return nil, ErrAgentNoReturn
		}

		if finish != nil {
			return finish.ReturnValues, nil
		}

		for _, action := range actions {
			tool, ok := nameToTool[action.Tool]
			if !ok {
				steps = append(steps, schema.AgentStep{
					Action:      action,
					Observation: fmt.Sprintf("%s is not a valid tool, try another one", action.Tool),
				})
			}

			observation, err := tool.Call(action.ToolInput)
			if err != nil {
				return nil, err
			}

			steps = append(steps, schema.AgentStep{
				Action:      action,
				Observation: observation,
			})
		}
	}

	return nil, ErrNotFinished
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

func (e Executor) GetMemory() schema.Memory {
	return memory.NewSimple()
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
