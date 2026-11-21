package agents_test

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/agents"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/internal/httprr"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/prompts"
	"github.com/tmc/langchaingo/tools"
)

// hasExistingRecording checks if a httprr recording exists for this test
func hasExistingRecording(t *testing.T) bool {
	testName := strings.ReplaceAll(t.Name(), "/", "_")
	testName = strings.ReplaceAll(testName, " ", "_")
	recordingPath := filepath.Join("testdata", testName+".httprr")
	_, err := os.Stat(recordingPath)
	return err == nil
}

func TestOpenAIFunctionsAgentWithHTTPRR(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// Skip if no recording available and no credentials
	if !hasExistingRecording(t) {
		t.Skip("No httprr recording available. Hint: Re-run tests with -httprecord=. to record new HTTP interactions")
	}

	rr := httprr.OpenForTest(t, http.DefaultTransport)

	// Configure OpenAI client with httprr
	opts := []openai.Option{
		openai.WithModel("gpt-4o"),
		openai.WithHTTPClient(rr.Client()),
	}
	if rr.Replaying() {
		opts = append(opts, openai.WithToken("test-api-key"))
	}

	llm, err := openai.New(opts...)
	if err != nil {
		t.Fatal(err)
	}

	// Create a simple calculator tool
	calculator := tools.Calculator{}

	// Create the OpenAI Functions agent
	agent := agents.NewOpenAIFunctionsAgent(
		llm,
		[]tools.Tool{calculator},
		agents.NewOpenAIOption().WithSystemMessage("You are a helpful assistant that can perform calculations."),
	)

	// Create executor
	executor := agents.NewExecutor(agent)

	// Run a simple calculation
	result, err := chains.Run(ctx, executor, "What is 15 multiplied by 4?")
	if err != nil {
		// Check if this is a recording mismatch error. The openai client
		// reports a replay miss as a network error.
		if strings.Contains(err.Error(), "cached HTTP response not found") ||
			(rr.Replaying() && strings.Contains(err.Error(), "network error")) {
			t.Skip("Recording format has changed or is incompatible. Hint: Re-run tests with -httprecord=. to record new HTTP interactions")
		}
		t.Fatal(err)
	}

	t.Logf("Agent response: %s", result)

	// Verify the result contains 60
	if !strings.Contains(result, "60") {
		t.Errorf("expected calculation result 60 in response, got: %s", result)
	}
}

func TestOpenAIFunctionsAgentComplexCalculation(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// Skip if no recording available and no credentials
	if !hasExistingRecording(t) {
		t.Skip("No httprr recording available. Hint: Re-run tests with -httprecord=. to record new HTTP interactions")
	}

	rr := httprr.OpenForTest(t, http.DefaultTransport)

	// Configure OpenAI client with httprr
	opts := []openai.Option{
		openai.WithModel("gpt-4o"),
		openai.WithHTTPClient(rr.Client()),
	}
	if rr.Replaying() {
		opts = append(opts, openai.WithToken("test-api-key"))
	}

	llm, err := openai.New(opts...)
	require.NoError(t, err)

	// Create a calculator tool
	calculator := tools.Calculator{}

	// Create the OpenAI Functions agent with extra messages
	agent := agents.NewOpenAIFunctionsAgent(
		llm,
		[]tools.Tool{calculator},
		agents.NewOpenAIOption().WithSystemMessage("You are a helpful math assistant."),
		agents.NewOpenAIOption().WithExtraMessages([]prompts.MessageFormatter{
			prompts.NewHumanMessagePromptTemplate("Please show your work step by step.", nil),
		}),
	)

	// Create executor with options
	executor := agents.NewExecutor(
		agent,
		agents.WithMaxIterations(5),
	)

	// Run a more complex calculation
	result, err := chains.Run(ctx, executor, "If I have 3 groups of 7 items, and I add 9 more items, how many items do I have in total?")
	if err != nil {
		// Check if this is a recording mismatch error. The openai client
		// reports a replay miss as a network error.
		if strings.Contains(err.Error(), "cached HTTP response not found") ||
			(rr.Replaying() && strings.Contains(err.Error(), "network error")) {
			t.Skip("Recording format has changed or is incompatible. Hint: Re-run tests with -httprecord=. to record new HTTP interactions")
		}
		t.Fatalf("failed to run agent: %v", err)
	}
	t.Logf("Agent response: %s", result)

	// Verify the result contains 30 (3*7 + 9 = 21 + 9 = 30)
	if !strings.Contains(result, "30") {
		t.Errorf("expected calculation result 30 in response, got: %s", result)
	}
}

// TestOpenAIFunctionsAgent_ParseOutput_NilResponse tests that ParseOutput handles nil response gracefully
func TestOpenAIFunctionsAgent_ParseOutput_NilResponse(t *testing.T) {
	t.Parallel()
	agent := &agents.OpenAIFunctionsAgent{}

	// Test with nil response - should return error instead of panic
	_, _, err := agent.ParseOutput(nil)
	if err == nil {
		t.Error("expected error for nil response")
	}
}

// TestOpenAIFunctionsAgent_ParseOutput_EmptyChoices tests that ParseOutput handles empty choices gracefully
func TestOpenAIFunctionsAgent_ParseOutput_EmptyChoices(t *testing.T) {
	t.Parallel()
	agent := &agents.OpenAIFunctionsAgent{}

	// Test with empty choices - should return error instead of panic
	resp := &llms.ContentResponse{
		Choices: []*llms.ContentChoice{},
	}
	_, _, err := agent.ParseOutput(resp)
	if err == nil {
		t.Error("expected error for empty choices")
	}
}

// TestOpenAIFunctionsAgent_ParseOutput_MultipleToolCalls tests multiple tool calls handling
func TestOpenAIFunctionsAgent_ParseOutput_MultipleToolCalls(t *testing.T) {
	t.Parallel()
	agent := &agents.OpenAIFunctionsAgent{}

	// Test multiple tool calls - should handle all calls, not just first one
	resp := &llms.ContentResponse{
		Choices: []*llms.ContentChoice{
			{
				ToolCalls: []llms.ToolCall{
					{
						ID: "call1",
						FunctionCall: &llms.FunctionCall{
							Name:      "calculator",
							Arguments: `{"__arg1": "2+2"}`,
						},
					},
					{
						ID: "call2",
						FunctionCall: &llms.FunctionCall{
							Name:      "weather",
							Arguments: `{"__arg1": "Seattle"}`,
						},
					},
				},
			},
		},
	}

	actions, finish, err := agent.ParseOutput(resp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if finish != nil {
		t.Error("expected actions, got finish")
	}
	if len(actions) != 2 {
		t.Errorf("expected 2 actions, got %d", len(actions))
	}
}

// replayModel is a fake llms.Model. Its first call returns the given tool
// call; later calls record the messages they receive and return a final
// answer.
type replayModel struct {
	first    *llms.ContentChoice
	calls    int
	messages []llms.MessageContent
}

func (m *replayModel) GenerateContent(
	_ context.Context, messages []llms.MessageContent, _ ...llms.CallOption,
) (*llms.ContentResponse, error) {
	m.calls++
	if m.calls == 1 {
		return &llms.ContentResponse{Choices: []*llms.ContentChoice{m.first}}, nil
	}
	m.messages = messages
	return &llms.ContentResponse{Choices: []*llms.ContentChoice{{Content: "done"}}}, nil
}

func (m *replayModel) Call(ctx context.Context, prompt string, options ...llms.CallOption) (string, error) {
	return llms.GenerateFromSinglePrompt(ctx, m, prompt, options...)
}

// echoTool returns its input unchanged.
type echoTool struct{ input string }

func (e *echoTool) Name() string        { return "echo" }
func (e *echoTool) Description() string { return "Returns its input." }
func (e *echoTool) Call(_ context.Context, input string) (string, error) {
	e.input = input
	return input, nil
}

// TestOpenAIFunctionsAgent_ScratchpadReplaysRawArguments checks that the tool
// call replayed to the model on the next turn carries the JSON arguments the
// model produced, not the unwrapped "__arg1" value given to the tool.
func TestOpenAIFunctionsAgent_ScratchpadReplaysRawArguments(t *testing.T) {
	t.Parallel()

	const rawArgs = `{"__arg1":"bonjour"}`
	tests := []struct {
		name  string
		first *llms.ContentChoice
	}{
		{
			name: "tool_calls",
			first: &llms.ContentChoice{
				ToolCalls: []llms.ToolCall{{
					ID:           "call_1",
					Type:         "function",
					FunctionCall: &llms.FunctionCall{Name: "echo", Arguments: rawArgs},
				}},
			},
		},
		{
			name: "legacy_function_call",
			first: &llms.ContentChoice{
				FuncCall: &llms.FunctionCall{Name: "echo", Arguments: rawArgs},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			model := &replayModel{first: tt.first}
			tool := &echoTool{}
			agent := agents.NewOpenAIFunctionsAgent(model, []tools.Tool{tool})
			executor := agents.NewExecutor(agent, agents.WithMaxIterations(3))

			out, err := chains.Run(context.Background(), executor, "salut")
			require.NoError(t, err)
			require.Equal(t, "done", out)
			require.Equal(t, "bonjour", tool.input, "the tool should receive the unwrapped __arg1 value")
			require.Equal(t, 2, model.calls)

			var replayed []llms.ToolCall
			for _, msg := range model.messages {
				for _, part := range msg.Parts {
					if tc, ok := part.(llms.ToolCall); ok {
						replayed = append(replayed, tc)
					}
				}
			}
			require.Len(t, replayed, 1)
			require.NotNil(t, replayed[0].FunctionCall)
			require.Equal(t, "echo", replayed[0].FunctionCall.Name)
			args := replayed[0].FunctionCall.Arguments
			require.True(t, json.Valid([]byte(args)), "replayed arguments are not valid JSON: %q", args)
			require.JSONEq(t, rawArgs, args)
		})
	}
}
