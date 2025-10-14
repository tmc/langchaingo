package agents_test

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vendasta/langchaingo/agents"
	"github.com/vendasta/langchaingo/chains"
	"github.com/vendasta/langchaingo/internal/httprr"
	"github.com/vendasta/langchaingo/llms/openai"
	"github.com/vendasta/langchaingo/prompts"
	"github.com/vendasta/langchaingo/schema"
	"github.com/vendasta/langchaingo/tools"
	"github.com/vendasta/langchaingo/tools/serpapi"
)

type testAgent struct {
	actions    []schema.AgentAction
	finish     *schema.AgentFinish
	err        error
	inputKeys  []string
	outputKeys []string
	tools      []tools.Tool

	recordedIntermediateSteps []schema.AgentStep
	recordedInputs            map[string]string
	numPlanCalls              int
}

func (a *testAgent) Plan(
	_ context.Context,
	intermediateSteps []schema.AgentStep,
	inputs map[string]string,
) ([]schema.AgentAction, *schema.AgentFinish, error) {
	a.recordedIntermediateSteps = intermediateSteps
	a.recordedInputs = inputs
	a.numPlanCalls++

	return a.actions, a.finish, a.err
}

func (a testAgent) GetInputKeys() []string {
	return a.inputKeys
}

func (a testAgent) GetOutputKeys() []string {
	return a.outputKeys
}

func (a *testAgent) GetTools() []tools.Tool {
	return a.tools
}

func TestExecutorWithErrorHandler(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	a := &testAgent{
		err: agents.ErrUnableToParseOutput,
	}
	executor := agents.NewExecutor(
		a,
		agents.WithMaxIterations(3),
		agents.WithParserErrorHandler(agents.NewParserErrorHandler(nil)),
	)

	_, err := chains.Call(ctx, executor, nil)
	require.ErrorIs(t, err, agents.ErrNotFinished)
	require.Equal(t, 3, a.numPlanCalls)
	require.Equal(t, []schema.AgentStep{
		{Observation: agents.ErrUnableToParseOutput.Error()},
		{Observation: agents.ErrUnableToParseOutput.Error()},
	}, a.recordedIntermediateSteps)
}

func TestExecutorWithMRKLAgent(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// Skip if no recording available and no credentials
	if !hasExistingRecording(t) {
		t.Skip("No httprr recording available. Hint: Re-run tests with -httprecord=. to record new HTTP interactions")
	}

	rr := httprr.OpenForTest(t, http.DefaultTransport)

	// Configure OpenAI client with httprr
	opts := []openai.Option{
		openai.WithModel("gpt-4"),
		openai.WithHTTPClient(rr.Client()),
	}
	if rr.Replaying() {
		opts = append(opts, openai.WithToken("test-api-key"))
	}

	llm, err := openai.New(opts...)
	require.NoError(t, err)

	serpapiOpts := []serpapi.Option{serpapi.WithHTTPClient(rr.Client())}
	if rr.Replaying() {
		serpapiOpts = append(serpapiOpts, serpapi.WithAPIKey("test-api-key"))
	}
	searchTool, err := serpapi.New(serpapiOpts...)
	require.NoError(t, err)

	calculator := tools.Calculator{}

	a, err := agents.Initialize(
		llm,
		[]tools.Tool{searchTool, calculator},
		agents.ZeroShotReactDescription,
	)
	require.NoError(t, err)

	result, err := chains.Run(ctx, a, "What is 5 plus 3? Please calculate this.") //nolint:lll
	if err != nil {
		// Check if this is a recording mismatch error
		if strings.Contains(err.Error(), "cached HTTP response not found") {
			t.Skip("Recording format has changed or is incompatible. Hint: Re-run tests with -httprecord=. to record new HTTP interactions")
		}
		require.NoError(t, err)
	}

	t.Logf("MRKL Agent response: %s", result)
	// Simple calculation: 5 + 3 = 8
	require.True(t, strings.Contains(result, "8"), "expected calculation result 8 in response")
}

func TestExecutorWithOpenAIFunctionAgent(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// Skip if no recording available and no credentials
	if !hasExistingRecording(t) {
		t.Skip("No httprr recording available. Hint: Re-run tests with -httprecord=. to record new HTTP interactions")
	}

	rr := httprr.OpenForTest(t, http.DefaultTransport)

	// Configure OpenAI client with httprr
	opts := []openai.Option{
		openai.WithModel("gpt-4"),
		openai.WithHTTPClient(rr.Client()),
	}
	if rr.Replaying() {
		opts = append(opts, openai.WithToken("test-api-key"))
	}

	llm, err := openai.New(opts...)
	require.NoError(t, err)

	serpapiOpts := []serpapi.Option{serpapi.WithHTTPClient(rr.Client())}
	if rr.Replaying() {
		serpapiOpts = append(serpapiOpts, serpapi.WithAPIKey("test-api-key"))
	}
	searchTool, err := serpapi.New(serpapiOpts...)
	require.NoError(t, err)

	calculator := tools.Calculator{}

	toolList := []tools.Tool{searchTool, calculator}

	a := agents.NewOpenAIFunctionsAgent(llm,
		toolList,
		agents.NewOpenAIOption().WithSystemMessage("you are a helpful assistant"),
		agents.NewOpenAIOption().WithExtraMessages([]prompts.MessageFormatter{
			prompts.NewHumanMessagePromptTemplate("please be strict", nil),
		}),
	)

	e := agents.NewExecutor(a)
	require.NoError(t, err)

	result, err := chains.Run(ctx, e, "when was the Go programming language tagged version 1.0?") //nolint:lll
	if err != nil {
		// Check if this is a recording mismatch error
		if strings.Contains(err.Error(), "cached HTTP response not found") {
			t.Skip("Recording format has changed or is incompatible. Hint: Re-run tests with -httprecord=. to record new HTTP interactions")
		}
		require.NoError(t, err)
	}

	t.Logf("Result: %s", result)

	require.True(t, strings.Contains(result, "2012") || strings.Contains(result, "March"),
		"correct answer 2012 or March not in response")
}

// mockTool implements the tools.Tool interface for testing
type mockTool struct {
	name             string
	description      string
	receivedInputPtr *string
}

func (m *mockTool) Name() string {
	return m.name
}

func (m *mockTool) Description() string {
	return m.description
}

func (m *mockTool) Call(_ context.Context, input string) (string, error) {
	*m.receivedInputPtr = input
	return "mock result", nil
}

func TestExecutorTrimsObservationSuffix(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// Create a mock tool that records what input it receives
	var receivedInput string
	mockToolInst := &mockTool{
		name:             "mock_tool",
		description:      "A mock tool for testing",
		receivedInputPtr: &receivedInput,
	}

	// Create a test agent that returns an action with trailing "\nObservation:"
	testAgent := &testAgent{
		actions: []schema.AgentAction{
			{
				Tool:      "mock_tool",
				ToolInput: "test input\nObservation:",
				Log:       "Action: mock_tool\nAction Input: test input\nObservation:",
			},
		},
		inputKeys:  []string{"input"},
		outputKeys: []string{"output"},
		tools:      []tools.Tool{mockToolInst},
	}

	executor := agents.NewExecutor(testAgent, agents.WithMaxIterations(1))

	_, err := chains.Call(ctx, executor, map[string]any{"input": "test question"})
	// We expect ErrNotFinished since our test agent doesn't provide a finish action
	require.ErrorIs(t, err, agents.ErrNotFinished)

	// Verify that the tool received the input with "\nObservation:" trimmed off
	require.Equal(t, "test input", receivedInput, "Tool should receive input with \\nObservation: suffix trimmed")
}
