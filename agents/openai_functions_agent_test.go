package agents_test

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/0xDezzy/langchaingo/agents"
	"github.com/0xDezzy/langchaingo/chains"
	"github.com/0xDezzy/langchaingo/internal/httprr"
	"github.com/0xDezzy/langchaingo/llms/openai"
	"github.com/0xDezzy/langchaingo/prompts"
	"github.com/0xDezzy/langchaingo/tools"
)

func TestOpenAIFunctionsAgentWithHTTPRR(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "OPENAI_API_KEY")

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

	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "OPENAI_API_KEY")

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
		t.Fatalf("failed to run agent: %v", err)
	}
	t.Logf("Agent response: %s", result)

	// Verify the result contains 30 (3*7 + 9 = 21 + 9 = 30)
	if !strings.Contains(result, "30") {
		t.Errorf("expected calculation result 30 in response, got: %s", result)
	}
}
