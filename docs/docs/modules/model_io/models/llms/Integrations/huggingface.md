---
title: Hugging Face
---

# Hugging Face

## Overview

This documentation provides a detailed overview and technical guidance for integrating the Hugging Face machine learning models with the LangchainGo library in the Go programming environment. This integration allows Go developers to leverage the power of pre-trained AI models for various applications, including natural language processing, text generation, and more.

## Prerequisites
Go programming language installed on your machine (version 1.15 or higher recommended).
A valid Hugging Face API token. Obtain it by creating an account on the Hugging Face platform and generating a new token

## Installation

```bash
go get github.com/tmc/langchaingo
```
Ensure that your Hugging Face API token is set as an environment variable:

```bash
export HUGGINGFACEHUB_API_TOKEN='your_hugging_face_api_token'
```

## Usage

[`examples/huggingface-llm-example/huggingface_example.go`](https://github.com/tmc/langchaingo/blob/main/examples/huggingface-llm-example/huggingface_example.go):

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/huggingface"
)

func main() {
	// You may instantiate a client with a custom token and/or custom model
	// clientOptions := []huggingface.Option{
	// 	huggingface.WithToken("HF_1234"),
	// 	huggingface.WithModel("ZZZ"),
	// }
	// llm, err := huggingface.New(clientOptions...)

	// Or you may instantiate a client with a default model and use token from environment variable
	llm, err := huggingface.New()
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()

	// Or override default model to another one
	generateOptions := []llms.CallOption{
		llms.WithModel("gpt2"),
		// llms.WithTopK(10),
		// llms.WithTopP(0.95),
		// llms.WithSeed(13),
	}
	completion, err := llms.GenerateFromSinglePrompt(
		ctx,
		llm,
		"What would be a good company name be for name a company that makes colorful socks?",
		generateOptions...)
	// Check for errors
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(completion)
}
```
