package agents_test

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/agents"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/internal/httprr"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/prompts"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/tools"
	"github.com/tmc/langchaingo/tools/serpapi"
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
	_ ...chains.ChainCallOption,
) ([]schema.AgentAction, *schema.AgentFinish, error) {
	a.recordedIntermediateSteps = intermediateSteps
	a.recordedInputs = inputs
	a.numPlanCalls++

	if a.numPlanCalls == 1 {
		return a.actions, nil, a.err
	}
	return nil, a.finish, a.err
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

func TestExecutorTrimsObservationSuffix(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// Create a mock tool that records what input it receives
	var receivedInput string
	mockToolInst := &mockTool{
		name:             "mock_tool",
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

type mockCallbackHandler struct {
	toolStartCalled bool
	toolEndCalled   bool
	toolErrorCalled bool
	toolStartInput  string
	toolEndInput    string
	toolErrorInput  error
}

func (m *mockCallbackHandler) HandleText(context.Context, string)                                   {}
func (m *mockCallbackHandler) HandleLLMStart(context.Context, []string)                             {}
func (m *mockCallbackHandler) HandleLLMGenerateContentStart(context.Context, []llms.MessageContent) {}
func (m *mockCallbackHandler) HandleLLMGenerateContentEnd(context.Context, *llms.ContentResponse)   {}
func (m *mockCallbackHandler) HandleLLMError(context.Context, error)                                {}
func (m *mockCallbackHandler) HandleChainStart(context.Context, map[string]any)                     {}
func (m *mockCallbackHandler) HandleChainEnd(context.Context, map[string]any)                       {}
func (m *mockCallbackHandler) HandleChainError(context.Context, error)                              {}
func (m *mockCallbackHandler) HandleToolStart(_ context.Context, input string) {
	m.toolStartCalled = true
	m.toolStartInput = input
}
func (m *mockCallbackHandler) HandleToolEnd(_ context.Context, output string) {
	m.toolEndCalled = true
	m.toolEndInput = output
}
func (m *mockCallbackHandler) HandleToolError(_ context.Context, err error) {
	m.toolErrorCalled = true
	m.toolErrorInput = err
}
func (m *mockCallbackHandler) HandleAgentAction(_ context.Context, _ schema.AgentAction)           {}
func (m *mockCallbackHandler) HandleAgentFinish(_ context.Context, _ schema.AgentFinish)           {}
func (m *mockCallbackHandler) HandleRetrieverStart(_ context.Context, _ string)                    {}
func (m *mockCallbackHandler) HandleRetrieverEnd(_ context.Context, _ string, _ []schema.Document) {}
func (m *mockCallbackHandler) HandleStreamingFunc(_ context.Context, _ []byte)                     {}

type mockTool struct {
	name             string
	response         string
	receivedInputPtr *string
	err              error
}

func (m mockTool) Name() string        { return m.name }
func (m mockTool) Description() string { return "A mock tool for testing" }
func (m mockTool) Call(_ context.Context, input string) (string, error) {
	if m.receivedInputPtr != nil {
		*m.receivedInputPtr = input
	}
	return m.response, m.err
}

func TestExecutorToolCallbacks(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		toolResponse  string
		toolError     error
		expectedStart string
		expectedEnd   string
		expectedError error
		shouldHaveErr bool
	}{
		{
			name:          "successful tool execution",
			toolResponse:  "success",
			expectedStart: "TEST_TOOL::(input)",
			expectedEnd:   "success",
			shouldHaveErr: false,
		},
		{
			name:          "tool execution with error",
			toolError:     fmt.Errorf("tool error"),
			expectedStart: "TEST_TOOL::(input)",
			expectedError: fmt.Errorf("tool error"),
			shouldHaveErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create mock tool and callback handler
			tool := mockTool{
				name:     "TEST_TOOL",
				response: tt.toolResponse,
				err:      tt.toolError,
			}
			callbackHandler := &mockCallbackHandler{}

			// Create test agent that will use our tool
			agent := &testAgent{
				actions: []schema.AgentAction{
					{
						Tool:      "TEST_TOOL",
						ToolInput: "input",
					},
				},
				tools: []tools.Tool{tool},
			}

			// Only set finish for successful tool execution
			if !tt.shouldHaveErr {
				agent.finish = &schema.AgentFinish{
					ReturnValues: map[string]any{
						"output": "done",
					},
				}
			}

			// Create executor with our callback handler
			executor := agents.NewExecutor(
				agent,
				agents.WithCallbacksHandler(callbackHandler),
			)

			// Execute the agent
			_, err := chains.Call(context.Background(), executor, map[string]any{"input": "test"})

			// Verify callback handling
			require.True(t, callbackHandler.toolStartCalled)
			require.Equal(t, tt.expectedStart, callbackHandler.toolStartInput)

			if tt.shouldHaveErr {
				require.True(t, callbackHandler.toolErrorCalled)
				require.Equal(t, tt.expectedError.Error(), callbackHandler.toolErrorInput.Error())
				require.False(t, callbackHandler.toolEndCalled)
				require.Error(t, err)
			} else {
				require.True(t, callbackHandler.toolEndCalled)
				require.Equal(t, tt.expectedEnd, callbackHandler.toolEndInput)
				require.False(t, callbackHandler.toolErrorCalled)
				require.NoError(t, err)
			}
		})
	}
}
