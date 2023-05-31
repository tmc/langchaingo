package main

import (
	"context"
	"fmt"
	"log"

	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/memory"
	"github.com/tmc/langchaingo/memory/history"
	"github.com/tmc/langchaingo/prompts"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	llm, err := openai.New()
	if err != nil {
		return err
	}

	chatHistory := history.NewSimpleChatMessageHistory()
	chainMemory := memory.NewBuffer(chatHistory)

	llmChain := chains.NewLLMChain(llm, prompts.NewPromptTemplate("How old is the planet {{.planet}}?", []string{"planet"}))
	llmChain.Memory = chainMemory

	ctx := context.Background()
	out, err := chains.Call(ctx, llmChain, map[string]any{"planet": "Earth"})
	if err != nil {
		return err
	}

	fmt.Println(out["text"])
	return nil
}
