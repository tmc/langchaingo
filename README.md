> 🎉 **Join our new official Discord community!** Connect with other LangChain Go developers, get help and contribute: [Join Discord](https://discord.gg/t9UbBQs2rG)

# 🦜️🔗 LangChain Go

[![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=flat-square)](https://pkg.go.dev/github.com/tmc/langchaingo)
[![scorecard](https://goreportcard.com/badge/github.com/tmc/langchaingo)](https://goreportcard.com/report/github.com/tmc/langchaingo)
[![](https://dcbadge.vercel.app/api/server/t9UbBQs2rG?compact=true&style=flat)](https://discord.gg/t9UbBQs2rG)
[![Open in Dev Containers](https://img.shields.io/static/v1?label=Dev%20Containers&message=Open&color=blue&logo=visualstudiocode)](https://vscode.dev/redirect?url=vscode://ms-vscode-remote.remote-containers/cloneInVolume?url=https://github.com/tmc/langchaingo)
[<img src="https://github.com/codespaces/badge.svg" title="Open in Github Codespace" width="150" height="20">](https://codespaces.new/tmc/langchaingo)

⚡ Building applications with LLMs through composability, with Go! ⚡

## 🤔 What is this?

This is the Go language implementation of [LangChain](https://github.com/langchain-ai/langchain).

## 📖 Documentation

- [Documentation Site](https://tmc.github.io/langchaingo/docs/)
- [API Reference](https://pkg.go.dev/github.com/tmc/langchaingo)


## 🎉 Examples

See [./examples](./examples) for example usage.

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
  ctx := context.Background()
  llm, err := openai.New()
  if err != nil {
    log.Fatal(err)
  }
  prompt := "What would be a good company name for a company that makes colorful socks?"
  completion, err := llms.GenerateFromSinglePrompt(ctx, llm, prompt)
  if err != nil {
    log.Fatal(err)
  }
  fmt.Println(completion)
}
```

```shell
$ go run .
Socktastic
```

# Resources

Join the Discord server for support and discussions: [Join Discord](https://discord.gg/8bHGKzHBkM)

Here are some links to blog posts and articles on using Langchain Go:

- [Using Gemini models in Go with LangChainGo](https://eli.thegreenplace.net/2024/using-gemini-models-in-go-with-langchaingo/) - Jan 2024
- [Using Ollama with LangChainGo](https://eli.thegreenplace.net/2023/using-ollama-with-langchaingo/) - Nov 2023
- [Creating a simple ChatGPT clone with Go](https://sausheong.com/creating-a-simple-chatgpt-clone-with-go-c40b4bec9267?sk=53a2bcf4ce3b0cfae1a4c26897c0deb0) - Aug 2023
- [Creating a ChatGPT Clone that Runs on Your Laptop with Go](https://sausheong.com/creating-a-chatgpt-clone-that-runs-on-your-laptop-with-go-bf9d41f1cf88?sk=05dc67b60fdac6effb1aca84dd2d654e) - Aug 2023


# Contributors

There is a momentum for moving the development of langchaingo to a more community effort, if you are interested in being a maintainer or you are a contributor please join our [Discord](https://discord.gg/8bHGKzHBkM) and let us know.

<a href="https://github.com/tmc/langchaingo/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=tmc/langchaingo" />
</a>
