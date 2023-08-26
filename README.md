# 🦜️🔗 LangChain Go

[![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=flat-square)](https://pkg.go.dev/github.com/tmc/langchaingo)
[![scorecard](https://goreportcard.com/badge/github.com/tmc/langchaingo)](https://goreportcard.com/report/github.com/tmc/langchaingo)

⚡ Building applications with LLMs through composability ⚡

## 🤔 What is this?

This is the Go language implementation of LangChain.

## 📖 Documentation

- [API Reference](https://pkg.go.dev/github.com/tmc/langchaingo)

## 🎉 Examples

See [./examples](./examples) for example usage.

```go
package main

import (
	"context"
	"log"

	"github.com/tmc/langchaingo/llms/openai"
)

func main() {
	llm, err := openai.New()
	if err != nil {
		log.Fatal(err)
	}
	prompt := "What would be a good company name for a company that makes colorful socks?"
	completion, err := llm.Call(context.Background(), prompt)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(completion)
}
```

```shell
$ go run .

Socktastic!
```

# Resources

Here are some links to blog posts and articles on using Langchain Go:

- [Creating a simple ChatGPT clone with Go](https://sausheong.com/creating-a-simple-chatgpt-clone-with-go-c40b4bec9267?sk=53a2bcf4ce3b0cfae1a4c26897c0deb0)
- [Creating a ChatGPT Clone that Runs on Your Laptop with Go](https://sausheong.com/creating-a-chatgpt-clone-that-runs-on-your-laptop-with-go-bf9d41f1cf88?sk=05dc67b60fdac6effb1aca84dd2d654e)