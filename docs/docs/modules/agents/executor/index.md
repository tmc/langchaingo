---
title: Agent Executors
---

# Agent Executors

<Info>
[Conceptual Guide](https://python.langchain.com/docs/modules/agents/concepts#agentexecutor)
</Info>

An executor runs the agent loop, invoking tools and returning their results to the agent until it produces an answer.

## Example
An example that initialize a MRKL (Modular Reasoning, Knowledge and Language, pronounced "miracle") agent executor.
[`examples/mrkl-agent-example/mrkl_agent.go`](https://github.com/tmc/langchaingo/blob/main/examples/mrkl-agent-example/mrkl_agent.go):

```go
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/tmc/langchaingo/agents"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/tools"
	"github.com/tmc/langchaingo/tools/serpapi"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	llm, err := openai.New()
	if err != nil {
		return err
	}
	search, err := serpapi.New()
	if err != nil {
		return err
	}
	agentTools := []tools.Tool{
		tools.Calculator{},
		search,
	}

	agent := agents.NewOneShotAgent(llm,
		agentTools,
		agents.WithMaxIterations(3))
	executor := agents.NewExecutor(agent)

	question := "Who is Olivia Wilde's boyfriend? What is his current age raised to the 0.23 power?"
	answer, err := chains.Run(context.Background(), executor, question)
	fmt.Println(answer)
	return err
}
```

- [Getting started](getting-started.md)
