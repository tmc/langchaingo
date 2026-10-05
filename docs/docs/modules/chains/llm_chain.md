---
title: "Getting Started: LLMChain"
draft: true
---

# Getting Started: LLMChain

<Info>
[Conceptual Guide](https://python.langchain.com/docs/modules/chains)
</Info>

An `LLMChain` is a simple chain that adds some functionality around language models. It is used widely throughout LangChain, including in other chains and agents.

An `LLMChain` consists of a `PromptTemplate` and a language model (either an [LLM](../model_io/models/llms/index.md) or [chat model](../model_io/models/chat/index.md)).

## Usage with LLMs

We can construct an LLMChain which takes user input, formats it with a PromptTemplate, and then passes the formatted response to an LLM:

[`examples/llm-chain-example/llm_chain.go`](https://github.com/tmc/langchaingo/blob/main/examples/llm-chain-example/llm_chain.go):

```go
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

	// If a chain only needs one input we can use Run to execute it.
	// We can pass callbacks to Run as an option, e.g:
	//   chains.WithCallback(callbacks.StreamLogHandler{})
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
```
