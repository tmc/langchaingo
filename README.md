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
