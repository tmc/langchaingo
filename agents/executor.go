package agents

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/tools"
)

const _intermediateStepsOutputKey = "intermediateSteps"

// Executor is the chain responsible for running agents.
type Executor struct {
	Agent            Agent
	Memory           schema.Memory
	CallbacksHandler callbacks.Handler
	ErrorHandler     *ParserErrorHandler

	MaxIterations           int
	ReturnIntermediateSteps bool
}

var (
	_ chains.Chain           = &Executor{}
	_ callbacks.HandlerHaver = &Executor{}
)

// NewExecutor creates a new agent executor with an agent and the tools the agent can use.
func NewExecutor(agent Agent, opts ...Option) *Executor {
	options := executorDefaultOptions()
	for _, opt := range opts {
		opt(&options)
	}

	return &Executor{
		Agent:                   agent,
		Memory:                  options.memory,
		MaxIterations:           options.maxIterations,
		ReturnIntermediateSteps: options.returnIntermediateSteps,
		CallbacksHandler:        options.callbacksHandler,
		ErrorHandler:            options.errorHandler,
	}
}

func (e *Executor) Call(ctx context.Context, inputValues map[string]any, _ ...chains.ChainCallOption) (map[string]any, error) { //nolint:lll
	// inputs, err := inputsToString(inputValues)
	// if err != nil {
	//	return nil, err
	//}
	nameToTool := getNameToTool(e.Agent.GetTools())

	steps := make([]schema.AgentStep, 0)
	var intermediateMessages []llms.ChatMessage
	var err error
	for i := 0; i < e.MaxIterations; i++ {
		var finish map[string]any
		steps, finish, intermediateMessages, err = e.doIteration(ctx, steps, nameToTool, inputValues, intermediateMessages)
		if finish != nil || err != nil {
			return finish, err
		}
	}

	if e.CallbacksHandler != nil {
		e.CallbacksHandler.HandleAgentFinish(ctx, schema.AgentFinish{
			ReturnValues: map[string]any{"output": ErrNotFinished.Error()},
		})
	}
	return e.getReturn(
		&schema.AgentFinish{ReturnValues: make(map[string]any)},
		steps,
	), ErrNotFinished
}

func (e *Executor) doIteration( // nolint
	ctx context.Context,
	steps []schema.AgentStep,
	nameToTool map[string]tools.Tool,
	inputs map[string]any,
	intermediateMessages []llms.ChatMessage,
) ([]schema.AgentStep, map[string]any, []llms.ChatMessage, error) {
	actions, finish, newIntermediateMessages, err := e.Agent.Plan(ctx, steps, inputs, intermediateMessages)
	if len(newIntermediateMessages) > 0 {
		intermediateMessages = append(intermediateMessages, newIntermediateMessages...)
	}
	if errors.Is(err, ErrUnableToParseOutput) && e.ErrorHandler != nil {
		formattedObservation := err.Error()
		if e.ErrorHandler.Formatter != nil {
			formattedObservation = e.ErrorHandler.Formatter(formattedObservation)
		}
		steps = append(steps, schema.AgentStep{
			Observation: formattedObservation,
		})
		return steps, nil, intermediateMessages, nil
	}
	if err != nil {
		return steps, nil, intermediateMessages, err
	}

	if len(actions) == 0 && finish == nil {
		return steps, nil, intermediateMessages, ErrAgentNoReturn
	}

	if finish != nil {
		if e.CallbacksHandler != nil {
			e.CallbacksHandler.HandleAgentFinish(ctx, *finish)
		}
		return steps, e.getReturn(finish, steps), intermediateMessages, nil
	}

	stepStreams := make([]<-chan schema.AgentStepWithError, len(actions))
	for index, action := range actions {
		stepStreams[index] = e.doAction(ctx, nameToTool, action)
	}
	for _, stepStream := range stepStreams {
		agentStepWithError := <-stepStream
		if agentStepWithError.Error != nil {
			return steps, nil, intermediateMessages, agentStepWithError.Error
		}
		steps = append(steps, agentStepWithError.AgentStep)
	}

	return steps, nil, intermediateMessages, nil
}

func (e *Executor) doAction(
	ctx context.Context,
	nameToTool map[string]tools.Tool,
	action schema.AgentAction,
) <-chan schema.AgentStepWithError {
	agentStepStream := make(chan schema.AgentStepWithError)
	go func() {
		defer close(agentStepStream)
		if e.CallbacksHandler != nil {
			e.CallbacksHandler.HandleAgentAction(ctx, action)
		}

		tool, ok := nameToTool[strings.ToUpper(action.Tool)]
		if !ok {
			agentStepStream <- schema.AgentStepWithError{
				AgentStep: schema.AgentStep{
					Action:      action,
					Observation: fmt.Sprintf("%s is not a valid tool, try another one", action.Tool),
				},
				Error: nil,
			}
			return
		}

		observation, err := tool.Call(ctx, action.ToolInput)
		if err != nil {
			agentStepStream <- schema.AgentStepWithError{
				AgentStep: schema.AgentStep{}, Error: err,
			}
			return
		}

		agentStepStream <- schema.AgentStepWithError{
			AgentStep: schema.AgentStep{
				Action:      action,
				Observation: observation,
			},
			Error: nil,
		}
	}()
	return agentStepStream
}

func (e *Executor) getReturn(finish *schema.AgentFinish, steps []schema.AgentStep) map[string]any {
	if e.ReturnIntermediateSteps {
		finish.ReturnValues[_intermediateStepsOutputKey] = steps
	}

	return finish.ReturnValues
}

// GetInputKeys gets the input keys the agent of the executor expects.
// Often "input".
func (e *Executor) GetInputKeys() []string {
	return e.Agent.GetInputKeys()
}

// GetOutputKeys gets the output keys the agent of the executor returns.
func (e *Executor) GetOutputKeys() []string {
	return e.Agent.GetOutputKeys()
}

func (e *Executor) GetMemory() schema.Memory { //nolint:ireturn
	return e.Memory
}

func (e *Executor) GetCallbackHandler() callbacks.Handler { //nolint:ireturn
	return e.CallbacksHandler
}

// func inputsToString(inputValues map[string]any) (map[string]string, error) {
//	inputs := make(map[string]string, len(inputValues))
//	for key, value := range inputValues {
//		valueStr, ok := value.(string)
//		if !ok {
//			return nil, fmt.Errorf("%w: %s", ErrExecutorInputNotString, key)
//		}
//
//		inputs[key] = valueStr
//	}
//
//	return inputs, nil
//}

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
