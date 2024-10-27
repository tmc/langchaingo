package main

import (
	"context"
	"fmt"
	"os"

	"github.com/starmvp/langchaingo/chains"
	"github.com/starmvp/langchaingo/llms/openai"
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
	llmMathChain := chains.NewLLMMathChain(llm)
	ctx := context.Background()
	out, err := chains.Run(ctx, llmMathChain, "What is 1024 plus six times 9?")
	fmt.Println(out)
	return err
}
