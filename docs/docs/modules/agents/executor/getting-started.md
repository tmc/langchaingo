---
title: "Getting Started: Agent Executors"
---

# Getting Started: Agent Executors

Agents use an LLM to determine which actions to take and in what order. An action can either be using a tool and observing its output, or returning to the user.

When used correctly agents can be extremely powerful. In this tutorial, we show you how to easily use agents through the simplest, highest level API.

In order to load agents, you should understand the following concepts:

- Tool: A function that performs a specific duty. This can be things like: Google Search, Database lookup, code REPL, other chains. The interface for a tool is currently a function that is expected to have a string as an input, with a string as an output.
- LLM: The language model powering the agent.
- Agent: The agent to use. This should be a string that references a support agent class. Because this notebook focuses on the simplest, highest level API, this only covers using the standard supported agents.

For this example, you'll need to set the SerpAPI environment variables in the `.env` file.

```bash
SERPAPI_API_KEY="..."
```

## Complete Example

Here's a complete working example showing how to create and run an agent executor:

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
