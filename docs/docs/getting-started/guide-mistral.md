---
title: "Quickstart: LangChainGo with Mistral"
---

# Quickstart: LangChainGo with Mistral

Get started by running your first program with LangChainGo and the [Mistral Platform](https://mistral.ai/).

## Prerequisites

1. **Mistral API Key**: Sign up on [Mistral](https://mistral.ai/) and retrieve your API key.
2. **Go**: [Download and install Go](https://go.dev/doc/install).

## Setup

Before interacting with the Mistral API, you need to set up your API key as an environment variable.

### Linux/macOS (bash/zsh)
```bash
export MISTRAL_API_KEY="your_mistral_api_key_here"
```

### Windows (Command Prompt)
```cmd
set MISTRAL_API_KEY=your_mistral_api_key_here
```

### Windows (PowerShell)
```powershell
$env:MISTRAL_API_KEY="your_mistral_api_key_here"
```

For permanent setup, add the environment variable to your shell's profile file (`~/.bashrc`, `~/.zshrc`, etc.) or system environment variables on Windows.

## Steps

1. **Set up your Mistral API Key**: Follow the setup instructions above to configure your API key.

2. **Run the example**: Execute the following command:

```shell
go run github.com/tmc/langchaingo/examples/mistral-completion-example@main
```

You should see output similar to the following:

```txt
The first man to walk on the moon was Neil Armstrong on July 20, 1969. He made this historic step during the Apollo 11 mission. Armstrong's famous quote upon setting foot on the lunar surface was, "That's one small step for man, one giant leap for mankind."
The first human to go to space was Yuri Gagarin, a Soviet cosmonaut. He completed an orbit around the Earth in the spacecraft Vostok 1 on April 12, 1961. This historic event marked the beginning of human space exploration.
```

The example sends prompts to Mistral and prints the responses.

Here is the entire program (from [mistral-completion-example](https://github.com/tmc/langchaingo/blob/main/examples/mistral-completion-example/mistral_completion_example.go)):

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/mistral"
)

func main() {
	llm, err := mistral.New(mistral.WithModel("open-mistral-7b"))
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()
	completionWithStreaming, err := llms.GenerateFromSinglePrompt(ctx, llm, "Who was the first man to walk on the moon?",
		llms.WithTemperature(0.8),
		llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
			fmt.Print(string(chunk))
			return nil
		}),
	)
	if err != nil {
		log.Fatal(err)
	}
	// The full string response will be available in completionWithStreaming after the streaming is complete.
	// (The Go compiler mandates declared variables be used at least once, hence the `_` assignment. https://go.dev/ref/spec#Blank_identifier)
	_ = completionWithStreaming

	completionWithoutStreaming, err := llms.GenerateFromSinglePrompt(ctx, llm, "Who was the first man to go to space?",
		llms.WithTemperature(0.2),
		llms.WithModel("mistral-small-latest"),
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("\n" + completionWithoutStreaming)
}
```
