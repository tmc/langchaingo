package agents

import (
	"context"
	"errors"
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

func (e *Executor) Call(ctx context.Context, inputValues map[string]any, options ...chains.ChainCallOption) (map[string]any, error) { //nolint:lll
	inputs, err := inputsToString(inputValues)
	if err != nil {
		return nil, err
	}
	nameToTool := getNameToTool(e.Agent.GetTools())

	steps := make([]schema.AgentStep, 0)
	for i := 0; i < e.MaxIterations; i++ {
		var finish map[string]any
		steps, finish, err = e.doIteration(ctx, steps, nameToTool, inputs, options...)
		if finish != nil || err != nil {
			return finish, err
		}

		// Check for potential infinite loops where the agent keeps repeating the same action
		if err := e.checkForLoop(steps); err != nil {
			if e.CallbacksHandler != nil {
				e.CallbacksHandler.HandleAgentFinish(ctx, schema.AgentFinish{
					ReturnValues: map[string]any{"output": err.Error()},
				})
			}
			return e.getReturn(
				&schema.AgentFinish{ReturnValues: make(map[string]any)},
				steps,
			), err
		}
	}

	if e.CallbacksHandler != nil {
		e.CallbacksHandler.HandleAgentFinish(ctx, schema.AgentFinish{
			ReturnValues: map[string]any{"output": ErrNotFinished.Error()},
		})
	}

	// Provide more helpful error message based on the steps taken
	err = e.buildNotFinishedError(steps)

	return e.getReturn(
		&schema.AgentFinish{ReturnValues: make(map[string]any)},
		steps,
	), err
}

func (e *Executor) doIteration( // nolint
	ctx context.Context,
	steps []schema.AgentStep,
	nameToTool map[string]tools.Tool,
	inputs map[string]string,
	options ...chains.ChainCallOption,
) ([]schema.AgentStep, map[string]any, error) {
	actions, finish, err := e.Agent.Plan(ctx, steps, inputs, options...)
	if errors.Is(err, ErrUnableToParseOutput) && e.ErrorHandler != nil {
		formattedObservation := err.Error()
		if e.ErrorHandler.Formatter != nil {
			formattedObservation = e.ErrorHandler.Formatter(formattedObservation)
		}
		steps = append(steps, schema.AgentStep{
			Observation: formattedObservation,
		})
		return steps, nil, nil
	}
	if err != nil {
		return steps, nil, err
	}

	if len(actions) == 0 && finish == nil {
		return steps, nil, ErrAgentNoReturn
	}

	if finish != nil {
		if e.CallbacksHandler != nil {
			e.CallbacksHandler.HandleAgentFinish(ctx, *finish)
		}
		return steps, e.getReturn(finish, steps), nil
	}

	for _, action := range actions {
		steps, err = e.doAction(ctx, steps, nameToTool, action, options...)
		if err != nil {
			return steps, nil, err
		}
	}

	return steps, nil, nil
}

func (e *Executor) doAction(
	ctx context.Context,
	steps []schema.AgentStep,
	nameToTool map[string]tools.Tool,
	action schema.AgentAction,
	options ...chains.ChainCallOption,
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

	// Call HandleToolStart before executing the tool
	toolInput := strings.TrimSuffix(action.ToolInput, "\nObservation:")
	if e.CallbacksHandler != nil {
		e.CallbacksHandler.HandleToolStart(ctx, toolInput)
	}

	observation, err := tool.Call(ctx, toolInput)
	if err != nil {
		// Call HandleToolError if tool execution fails
		if e.CallbacksHandler != nil {
			e.CallbacksHandler.HandleToolError(ctx, err)
		}
		return nil, err
	}

	// Call HandleToolEnd after successful tool execution
	if e.CallbacksHandler != nil {
		e.CallbacksHandler.HandleToolEnd(ctx, observation)
	}

	return append(steps, schema.AgentStep{
		Action:      action,
		Observation: observation,
	}), nil
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

// checkForLoop detects if the agent is stuck in a loop by checking if the same
// tool is being called repeatedly with the same input.
func (e *Executor) checkForLoop(steps []schema.AgentStep) error {
	if len(steps) < 3 {
		return nil
	}

	// Check the last 3 steps for identical tool calls
	lastStep := steps[len(steps)-1]
	if lastStep.Action.Tool == "" {
		return nil // Skip steps without actions (e.g., parser errors)
	}

	identicalCount := 1
	for i := len(steps) - 2; i >= 0 && i >= len(steps)-3; i-- {
		step := steps[i]
		if step.Action.Tool == lastStep.Action.Tool &&
			step.Action.ToolInput == lastStep.Action.ToolInput {
			identicalCount++
		}
	}

	// If the same tool with the same input was called 3 times in a row, it's likely a loop
	if identicalCount >= 3 {
		return fmt.Errorf("%w: agent is repeating the same action (%s with input %q). This usually means the LLM is not generating a 'Final Answer'. Consider using a different model or adjusting the prompt",
			ErrNotFinished,
			lastStep.Action.Tool,
			lastStep.Action.ToolInput)
	}

	return nil
}

// buildNotFinishedError creates a more helpful error message when the agent doesn't finish.
func (e *Executor) buildNotFinishedError(steps []schema.AgentStep) error {
	if len(steps) == 0 {
		return fmt.Errorf("%w: agent took no actions. Check if the LLM is generating valid output", ErrNotFinished)
	}

	// Count how many steps had actions vs parser errors
	actionCount := 0
	parseErrorCount := 0
	for _, step := range steps {
		if step.Action.Tool != "" {
			actionCount++
		} else if strings.Contains(step.Observation, "unable to parse") {
			parseErrorCount++
		}
	}

	if parseErrorCount > 0 {
		return fmt.Errorf("%w: agent failed to parse LLM output %d/%d times. The LLM may not be following the expected format. Consider using a different model or adjusting the prompt",
			ErrNotFinished, parseErrorCount, len(steps))
	}

	if actionCount > 0 {
		lastAction := ""
		for i := len(steps) - 1; i >= 0; i-- {
			if steps[i].Action.Tool != "" {
				lastAction = steps[i].Action.Tool
				break
			}
		}
		return fmt.Errorf("%w: agent executed %d actions (last: %s) but never generated a 'Final Answer'. This usually means the LLM is not instruction-following well. Consider using a more capable model",
			ErrNotFinished, actionCount, lastAction)
	}

	return ErrNotFinished
}
