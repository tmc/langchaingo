package agents_test

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/agents"
	"github.com/tmc/langchaingo/chains"
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

	a := &testAgent{
		err: agents.ErrUnableToParseOutput,
	}
	executor := agents.NewExecutor(
		a,
		agents.WithMaxIterations(3),
		agents.WithParserErrorHandler(agents.NewParserErrorHandler(nil)),
	)

	_, err := chains.Call(context.Background(), executor, nil)
	require.ErrorIs(t, err, agents.ErrNotFinished)
	require.Equal(t, 3, a.numPlanCalls)
	require.Equal(t, []schema.AgentStep{
		{Observation: agents.ErrUnableToParseOutput.Error()},
		{Observation: agents.ErrUnableToParseOutput.Error()},
	}, a.recordedIntermediateSteps)
}

func TestExecutorWithMRKLAgent(t *testing.T) {
	t.Parallel()

	if openaiKey := os.Getenv("OPENAI_API_KEY"); openaiKey == "" {
		t.Skip("OPENAI_API_KEY not set")
	}
	if serpapiKey := os.Getenv("SERPAPI_API_KEY"); serpapiKey == "" {
		t.Skip("SERPAPI_API_KEY not set")
	}

	llm, err := openai.New()
	require.NoError(t, err)

	searchTool, err := serpapi.New()
	require.NoError(t, err)

	calculator := tools.Calculator{}

	a, err := agents.Initialize(
		llm,
		[]tools.Tool{searchTool, calculator},
		agents.ZeroShotReactDescription,
	)
	require.NoError(t, err)

	result, err := chains.Run(context.Background(), a, "If a person lived three times as long as Jacklyn Zeman, how long would they live") //nolint:lll
	require.NoError(t, err)

	require.True(t, strings.Contains(result, "210"), "correct answer 210 not in response")
}

func TestExecutorWithOpenAIFunctionAgent(t *testing.T) {
	t.Parallel()

	if openaiKey := os.Getenv("OPENAI_API_KEY"); openaiKey == "" {
		t.Skip("OPENAI_API_KEY not set")
	}
	if serpapiKey := os.Getenv("SERPAPI_API_KEY"); serpapiKey == "" {
		t.Skip("SERPAPI_API_KEY not set")
	}

	llm, err := openai.New()
	require.NoError(t, err)

	searchTool, err := serpapi.New()
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

	result, err := chains.Run(context.Background(), e, "what is HK singer Eason Chan's years old?") //nolint:lll
	require.NoError(t, err)

	require.True(t, strings.Contains(result, "47") || strings.Contains(result, "49"),
		"correct answer 47 or 49 not in response")
}

type mockCallbackHandler struct {
	toolStartCalled bool
	toolEndCalled   bool
	toolErrorCalled bool
	toolStartInput  string
	toolEndInput    string
	toolErrorInput  error
}

func (m *mockCallbackHandler) HandleText(_ context.Context, _ string)       {}
func (m *mockCallbackHandler) HandleLLMStart(_ context.Context, _ []string) {}
func (m *mockCallbackHandler) HandleLLMGenerateContentStart(_ context.Context, _ []llms.MessageContent) {
}
func (m *mockCallbackHandler) HandleLLMGenerateContentEnd(_ context.Context, _ *llms.ContentResponse) {
}
func (m *mockCallbackHandler) HandleLLMError(_ context.Context, _ error)            {}
func (m *mockCallbackHandler) HandleChainStart(_ context.Context, _ map[string]any) {}
func (m *mockCallbackHandler) HandleChainEnd(_ context.Context, _ map[string]any)   {}
func (m *mockCallbackHandler) HandleChainError(_ context.Context, _ error)          {}
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
	name     string
	response string
	err      error
}

func (m mockTool) Name() string                                     { return m.name }
func (m mockTool) Description() string                              { return "A mock tool for testing" }
func (m mockTool) Call(_ context.Context, _ string) (string, error) { return m.response, m.err }

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
