---
title: watsonx
---

# watsonx

Integration support for [IBM watsonx](https://www.ibm.com/watsonx) foundation models with [`watsonx-go`](https://github.com/IBM/watsonx-go
).

## Setup

You will need to set the following environment variables for using the WatsonX AI API.

- `WATSONX_API_KEY`: generate from your [IBM Cloud account](https://cloud.ibm.com/iam/apikeys).
- `WATSONX_PROJECT_ID`: copy from your [watsonx project settings](https://dataplatform.cloud.ibm.com/projects/?context=wx).

Alternatively, these can be passed into the model on creation:

```go
import (
	wx "github.com/IBM/watsonx-go/pkg/models"
    "github.com/tmc/langchaingo/llms/watsonx"
)
...
llm, _ := watsonx.New(
    wx.WithWatsonxAPIKey("YOUR WATSONX API KEY"),
    wx.WithWatsonxProjectID("YOUR WATSONX PROJECT ID"),
)
```

## Example

[`examples/watsonx-llm-example/watsonx_example.go`](https://github.com/tmc/langchaingo/blob/main/examples/watsonx-llm-example/watsonx_example.go):

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/watsonx"
)

func main() {
	llm, err := watsonx.New(
		"meta-llama/llama-3-70b-instruct",
		//// Optional parameters:
		// wx.WithWatsonxAPIKey("YOUR WATSONX API KEY"),
		// wx.WithWatsonxProjectID("YOUR WATSONX PROJECT ID"),
	)
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()

	// Or override default model to another one
	prompt := "What would be a good company name be for name a company that makes colorful socks?"
	completion, err := llms.GenerateFromSinglePrompt(
		ctx,
		llm,
		prompt,
		llms.WithTopK(10),
		llms.WithTopP(0.95),
		llms.WithSeed(13),
	)
	// Check for errors
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(completion)
}
```
