
# Perplexity Tool Integration for Agents

Use perplexity in your AI Agent to enrich it with data from the web.

Full code example:

```go
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/tmc/langchaingo/agents"
	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/tools"
	"github.com/tmc/langchaingo/tools/perplexity"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	llm, err := openai.New(
		openai.WithModel("gpt-4o-mini"),
		openai.WithCallback(callbacks.LogHandler{}),
	)
	if err != nil {
		return err
	}

	perpl, err := perplexity.NewPerplexity(perplexity.ModelLlamaSonarSmall)
	if err != nil {
		return err
	}

	agentTools := []tools.Tool{
		perpl,
	}

	agent := agents.NewOneShotAgent(llm,
		agentTools,
		agents.WithMaxIterations(2),
	)
	executor := agents.NewExecutor(agent)

	question := "what's the latest and best LLM on the market at the moment?"
	answer, err := chains.Run(context.Background(), executor, question)

	fmt.Println(answer)

	return err
}
```