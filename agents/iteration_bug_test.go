package agents_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/agents"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/tools"
)

// loopingAgent simulates an agent that keeps returning actions without finishing
// This reproduces the issue from #1225
type loopingAgent struct {
	callCount  int
	tools      []tools.Tool
	inputKeys  []string
	outputKeys []string
}

func (a *loopingAgent) Plan(
	_ context.Context,
	intermediateSteps []schema.AgentStep,
	inputs map[string]string,
	_ ...chains.ChainCallOption,
) ([]schema.AgentAction, *schema.AgentFinish, error) {
	a.callCount++

	// Simulate what's happening in issue #1225:
	// The agent keeps generating new actions instead of finishing
	// even after tool calls succeed

	// Return a new action each time, simulating the problematic behavior
	// where models like llama3.2 don't output "Final Answer:"
	return []schema.AgentAction{
		{
			Tool:      "calculator",
			ToolInput: fmt.Sprintf("%d + %d", a.callCount, a.callCount),
			Log:       fmt.Sprintf("Thought: I need to calculate %d + %d\nAction: calculator\nAction Input: %d + %d", a.callCount, a.callCount, a.callCount, a.callCount),
		},
	}, nil, nil
}

func (a *loopingAgent) GetInputKeys() []string {
	return a.inputKeys
}

func (a *loopingAgent) GetOutputKeys() []string {
	return a.outputKeys
}

func (a *loopingAgent) GetTools() []tools.Tool {
	return a.tools
}

// TestExecutorIterationCounting verifies that the executor correctly counts iterations
// and stops after MaxIterations
func TestExecutorIterationCounting(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	agent := &loopingAgent{
		inputKeys:  []string{"input"},
		outputKeys: []string{"output"},
		tools:      []tools.Tool{tools.Calculator{}},
	}

	maxIter := 5
	executor := agents.NewExecutor(agent, agents.WithMaxIterations(maxIter))

	_, err := chains.Run(ctx, executor, "test")

	// Should get ErrNotFinished since the agent never returns a finish
	require.ErrorIs(t, err, agents.ErrNotFinished)

	// The agent's Plan method should be called exactly maxIter times
	require.Equal(t, maxIter, agent.callCount,
		"Plan should be called exactly MaxIterations times when agent doesn't finish")
}

// finishingAgent is an agent that finishes after a specific number of calls
type finishingAgent struct {
	callCount  int
	finishAt   int
	tools      []tools.Tool
	inputKeys  []string
	outputKeys []string
}

func (a *finishingAgent) Plan(
	_ context.Context,
	intermediateSteps []schema.AgentStep,
	inputs map[string]string,
	_ ...chains.ChainCallOption,
) ([]schema.AgentAction, *schema.AgentFinish, error) {
	a.callCount++

	if a.callCount >= a.finishAt {
		return nil, &schema.AgentFinish{
			ReturnValues: map[string]any{"output": "done"},
			Log:          "Final Answer: done",
		}, nil
	}

	return []schema.AgentAction{
		{
			Tool:      "calculator",
			ToolInput: "1 + 1",
			Log:       "Action: calculator\nAction Input: 1 + 1",
		},
	}, nil, nil
}

func (a *finishingAgent) GetInputKeys() []string {
	return a.inputKeys
}

func (a *finishingAgent) GetOutputKeys() []string {
	return a.outputKeys
}

func (a *finishingAgent) GetTools() []tools.Tool {
	return a.tools
}

// TestExecutorEarlyFinish verifies that the executor stops early when the agent returns a finish
func TestExecutorEarlyFinish(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	finishAfter := 2
	agent := &finishingAgent{
		finishAt:   finishAfter,
		inputKeys:  []string{"input"},
		outputKeys: []string{"output"},
		tools:      []tools.Tool{tools.Calculator{}},
	}

	executor := agents.NewExecutor(agent, agents.WithMaxIterations(10))

	result, err := chains.Run(ctx, executor, "test")

	require.NoError(t, err)
	require.Equal(t, "done", result)
	require.Equal(t, finishAfter, agent.callCount,
		"Plan should be called exactly finishAfter times when agent finishes early")
}

// repeatingAgent simulates an agent that keeps calling the same tool with the same input
// This should trigger loop detection
type repeatingAgent struct {
	callCount  int
	tools      []tools.Tool
	inputKeys  []string
	outputKeys []string
}

func (a *repeatingAgent) Plan(
	_ context.Context,
	intermediateSteps []schema.AgentStep,
	inputs map[string]string,
	_ ...chains.ChainCallOption,
) ([]schema.AgentAction, *schema.AgentFinish, error) {
	a.callCount++

	// Always return the same action (infinite loop)
	return []schema.AgentAction{
		{
			Tool:      "calculator",
			ToolInput: "1 + 1",
			Log:       "Action: calculator\nAction Input: 1 + 1",
		},
	}, nil, nil
}

func (a *repeatingAgent) GetInputKeys() []string {
	return a.inputKeys
}

func (a *repeatingAgent) GetOutputKeys() []string {
	return a.outputKeys
}

func (a *repeatingAgent) GetTools() []tools.Tool {
	return a.tools
}

// TestExecutorLoopDetection verifies that the executor detects when an agent is stuck in a loop
func TestExecutorLoopDetection(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	agent := &repeatingAgent{
		inputKeys:  []string{"input"},
		outputKeys: []string{"output"},
		tools:      []tools.Tool{tools.Calculator{}},
	}

	executor := agents.NewExecutor(agent, agents.WithMaxIterations(10))

	_, err := chains.Run(ctx, executor, "test")

	// Should detect the loop and return an error
	require.Error(t, err)
	require.ErrorIs(t, err, agents.ErrNotFinished)
	require.Contains(t, err.Error(), "repeating the same action")
	require.Contains(t, err.Error(), "calculator")

	// Should detect the loop after 3 identical calls
	require.Equal(t, 3, agent.callCount, "Loop should be detected after 3 identical calls")
}

// TestExecutorBetterErrorMessages verifies that the executor provides helpful error messages
func TestExecutorBetterErrorMessages(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	tests := []struct {
		name          string
		agent         agents.Agent
		expectedError string
		maxIterations int
		expectedCalls int
	}{
		{
			name: "agent with varying actions",
			agent: &loopingAgent{
				inputKeys:  []string{"input"},
				outputKeys: []string{"output"},
				tools:      []tools.Tool{tools.Calculator{}},
			},
			expectedError: "never generated a 'Final Answer'",
			maxIterations: 5,
			expectedCalls: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor := agents.NewExecutor(tt.agent, agents.WithMaxIterations(tt.maxIterations))

			_, err := chains.Run(ctx, executor, "test")

			require.Error(t, err)
			require.ErrorIs(t, err, agents.ErrNotFinished)
			require.Contains(t, err.Error(), tt.expectedError)
		})
	}
}
