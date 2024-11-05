package perplexity

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/tmc/langchaingo/agents"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/tools"
)

func TestRun(t *testing.T) {
	t.Parallel()

	if os.Getenv("PERPLEXITY_API_KEY") == "" {
		t.Skip("PERPLEXITY_API_KEY not set")
	}
	if os.Getenv("OPENAI_API_KEY") == "" {
		t.Skip("OPENAI_API_KEY not set")
	}

	llm, err := openai.New()
	if err != nil {
		t.Fatalf("failed to create LLM: %v", err)
	}

	perpl, err := NewPerplexity(ModelLlamaSonarSmall)
	if err != nil {
		t.Fatalf("failed to create Perplexity tool: %v", err)
	}

	agentTools := []tools.Tool{
		perpl,
	}

	agent := agents.NewOneShotAgent(llm,
		agentTools,
		agents.WithMaxIterations(1),
	)
	executor := agents.NewExecutor(agent)

	question := "what is the largest country in the world by total area?"
	answer, err := chains.Run(context.Background(), executor, question)
	if err != nil {
		t.Fatalf("failed to run chains: %v", err)
	}

	const expectedAnswer = "Russia"
	if !strings.Contains(answer, expectedAnswer) {
		t.Errorf("expected answer to contain %q, got %q", expectedAnswer, answer)
	}
}
