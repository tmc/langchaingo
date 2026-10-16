package main

import (
	"context"
	"encoding/json"
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

	translatePrompt := prompts.NewPromptTemplate(
		"Translate the following text from {{.inputLanguage}} to {{.outputLanguage}}. {{.text}}",
		[]string{"inputLanguage", "outputLanguage", "text"},
	)
	llmChain := chains.NewLLMChain(llm, translatePrompt)
	llmChain.EnableMultiPrompt()

	// Otherwise the call function must be used.
	outputValues, err := chains.Call(context.Background(), llmChain, map[string]any{
		"inputLanguage":  "English",
		"outputLanguage": "French",
		"text":           "I love programming.",
	})
	if err != nil {
		return err
	}

	out, ok := outputValues[llmChain.OutputKey]
	if !ok {
		return fmt.Errorf("invalid chain return")
	}
	cnt, err := json.Marshal(out)
	if err != nil {
		return err
	}
	fmt.Println(string(cnt))

	return nil
}
