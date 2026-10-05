---
title: OpenAI
---

# OpenAI
OpenAI offers a spectrum of models with different levels of power suitable for different tasks.

This example goes over how to use LangChain to interact with OpenAI models.

There are two options to set the the OpenAI key. 

1. We can do this by setting the environment variable `OPENAI_API_KEY` to the API key.

2. Or we can do it when initializing the wrapper along with other arguments:

```go
model, err := openai.New(openai.WithToken(apiToken))
```

## Example

[`examples/openai-completion-example/openai_completion_example.go`](https://github.com/tmc/langchaingo/blob/main/examples/openai-completion-example/openai_completion_example.go):

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)

func main() {
	llm, err := openai.New()
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()
	completion, err := llms.GenerateFromSinglePrompt(ctx,
		llm,
		"The first man to walk on the moon",
		llms.WithTemperature(0.8),
		llms.WithStopWords([]string{"Armstrong"}),
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("The first man to walk on the moon:")
	fmt.Println(completion)
}
```
