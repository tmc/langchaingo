# 🦜️🔗 LangChain Go

[![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=flat-square)](https://pkg.go.dev/github.com/tmc/langchaingo)

⚡ Building applications with LLMs through composability ⚡

## 🤔 What is this?

This is the Go language implementation of LangChain.

## 📖 Documentation

- [API Reference](https://pkg.go.dev/github.com/tmc/langchaingo)

## 🎉 Examples

See [./examples](./examples) for example usage.

```go
import (
	"github.com/tmc/langchaingo/llms/openai"
)

func main() {
	llm, err := openai.New()
	if err != nil {
		log.Fatal(err)
	}
	completion, err := llm.Call("What would be a good company name a company that makes colorful socks?")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(completion)
}
```
```shell
$ go run .

Socktastic!
```
