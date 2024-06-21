package main

import (
	"context"
	"fmt"
	"github.com/tmc/langchaingo/callbacks"
	"log"
	"os"

	"github.com/tmc/langchaingo/agents"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/tools"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	llm, err := openai.New(openai.WithModel("gpt-4-turbo"))
	if err != nil {
		log.Fatal(err)
	}

	agentTools := []tools.Tool{
		tools.Calculator{},
	}

	agent := agents.NewOpenAIFunctionsAgent(
		llm,
		agentTools,
		agents.WithCallbacksHandler(callbacks.LogHandler{}),
	)
	if err != nil {
		return err
	}

	executor := agents.NewExecutor(agent,
		agents.WithMaxIterations(3),
		agents.WithReturnIntermediateSteps(),
	)

	a, err := executor.Call(context.Background(), map[string]any{"input": "what is 3 plus 3 and what is python"})
	fmt.Println(a, err)
	return err
}
