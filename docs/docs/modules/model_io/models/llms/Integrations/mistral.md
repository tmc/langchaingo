---
title: Mistral
---

# Mistral
The Mistral Platform (https://mistral.ai) offers a spectrum of models with different levels of power suitable for different tasks.

This example goes over how to use LangChain to interact with Mistral models, with an example that shows how to get a streaming and non-streaming completion from the Mistral API using the LangChainGo wrapper.

## Configuring the API key
There are two options to set the the Mistral Platform API key.

1. We can do this by setting the environment variable `MISTRAL_API_KEY` to the api key.

2. Or we can do it when initializing the wrapper along with other arguments.
   ```go
    model, err := mistral.New(mistral.WithAPIKey(apiKey))
   ```

## Setting the model name
As mentioned above, there are many models available on the Mistral platform. We can set the model name when initializing the wrapper, and it can be overridden when performing completions through `langchaingo`.
```go
    model, err := mistral.New(mistral.WithAPIKey(apiKey), mistral.WithModel("mistral-small-latest"))
```
  * Currently-listed models on Mistral.ai:
  * `open-mistral-7b` (aka mistral-tiny-2312)
  * `open-mixtral-8x7b` (aka mistral-small-2312): `Note:` DOES NOT seem usable via the mistral-go client library at the moment.
  * `mistral-small-latest` (aka mistral-small-2402)
  * `mistral-medium-latest` (aka mistral-medium-2312)
  * `mistral-large-latest` (aka mistral-large-2402)

[`examples/mistral-completion-example/mistral_completion_example.go`](https://github.com/tmc/langchaingo/blob/main/examples/mistral-completion-example/mistral_completion_example.go):

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
