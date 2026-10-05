---
title: Groq
---

# Groq

## Overview

Use Groq models from a Go application with LangChainGo.

## Prerequisites

- Go programming language installed on your machine (version 1.22.0 or higher recommended).
-  A valid Groq API key. Obtain it by creating an account on the Groq platform and generating a new token.

## Installation

To install the Groq package in your Go project, run the following command:

```bash
go get github.com/tmc/langchaingo
```

Set the API key in your shell before running the example:

```bash
export GROQ_API_KEY="your-api-key"
```

The example reads `GROQ_API_KEY` from the process environment. A `.env`
file requires a separate loader or shell setup.

## Usage

[`examples/groq-completion-example/groq_completion_example.go`](https://github.com/tmc/langchaingo/blob/main/examples/groq-completion-example/groq_completion_example.go):

```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)

func main() {
	apiKey := os.Getenv("GROQ_API_KEY")

	llm, err := openai.New(
		openai.WithModel("moonshotai/kimi-k2-instruct"),
		openai.WithBaseURL("https://api.groq.com/openai/v1"),
		openai.WithToken(apiKey),
	)
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()
	_, err = llms.GenerateFromSinglePrompt(ctx,
		llm,
		"Write a long poem about how golang is a fantastic language.",
		llms.WithTemperature(0.8),
		llms.WithMaxTokens(4096),
		llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
			fmt.Print(string(chunk))
			return nil
		}),
	)
	fmt.Println()
	if err != nil {
		log.Fatal(err)
	}
}
```
