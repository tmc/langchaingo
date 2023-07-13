package main

import (
	"context"
	"log"

	"github.com/tmc/langchaingo/agents"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/memory"
	"github.com/tmc/langchaingo/tools"
	"github.com/tmc/langchaingo/tools/serpapi"
)

func main() {
	postgreAdapter, err := NewPostgreAdapter()
	if err != nil {
		log.Print(err)
	}

	persistentMemoryBuffer := memory.NewPersistentBuffer(postgreAdapter)
	persistentMemoryBuffer.DB.SetSessionID("USID-001")

	llm, err := openai.New()
	if err != nil {
		log.Print(err)
	}

	serpapi, err := serpapi.New()
	if err != nil {
		log.Print(err)
	}

	iterations := 3

	executor, err := agents.Initialize(
		llm,
		[]tools.Tool{serpapi},
		agents.ZeroShotReactDescription,
		agents.WithMemory(persistentMemoryBuffer),
		agents.WithMaxIterations(iterations),
	)
	if err != nil {
		log.Print(err)
	}

	input := "Who is the current CEO of Twitter?"
	answer, err := chains.Run(context.Background(), executor, input)
	if err != nil {
		log.Print(err)
		return
	}

	log.Print(answer)
}
