package main

import (
	"context"
	"fmt"
	"os"

	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/prompts"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	// We can construct an LLMChain from a PromptTemplate and an LLM.
	llm, err := openai.New()
	if err != nil {
		return err
	}
	prompt := prompts.NewPromptTemplate(
		"What is a good name for a company that makes {{.product}}?",
		[]string{"product"},
	)
	llmChain := chains.NewLLMChain(llm, prompt)

	// If a chain only needs one input we can use the run function to execute chain.
	ctx := context.Background()
	out, err := chains.Run(ctx, llmChain, "socks")
	if err != nil {
		return err
	}
	fmt.Println(out)

	translatePrompt := prompts.NewPromptTemplate(
		"Translate the following text from {{.inputLanguage}} to {{.outputLanguage}}. {{.text}}",
		[]string{"inputLanguage", "outputLanguage", "text"},
	)
	llmChain = chains.NewLLMChain(llm, translatePrompt)

	// Otherwise the call function must be used.
	outputValues, err := chains.Call(ctx, llmChain, map[string]any{
		"inputLanguage":  "English",
		"outputLanguage": "French",
		"text":           "I love programming.",
	})
	if err != nil {
		return err
	}

	out, ok := outputValues[llmChain.OutputKey].(string)
	if !ok {
		return fmt.Errorf("invalid chain return")
	}
	fmt.Println(out)

	return nil
}
