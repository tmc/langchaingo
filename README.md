> 🎉 **加入我们全新的官方 Discord 社区！** 与其他 LangChain Go 开发者交流、获取帮助并做出贡献：[加入 Discord](https://discord.gg/t9UbBQs2rG)

# 🦜️🔗 LangChain Go

[![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=flat-square)](https://pkg.go.dev/github.com/tmc/langchaingo)
[![scorecard](https://goreportcard.com/badge/github.com/tmc/langchaingo)](https://goreportcard.com/report/github.com/tmc/langchaingo)
[![](https://dcbadge.vercel.app/api/server/t9UbBQs2rG?compact=true&style=flat)](https://discord.gg/t9UbBQs2rG)
[![在 Dev Containers 中打开](https://img.shields.io/static/v1?label=Dev%20Containers&message=Open&color=blue&logo=visualstudiocode)](https://vscode.dev/redirect?url=vscode://ms-vscode-remote.remote-containers/cloneInVolume?url=https://github.com/tmc/langchaingo)
[<img src="https://github.com/codespaces/badge.svg" title="在 Github Codespace 中打开" width="150" height="20">](https://codespaces.new/tmc/langchaingo)

⚡ 通过可组合性，使用 Go 构建 LLM (大型语言模型) 应用程序！ ⚡

## 🤔 这是什么？

这是 [LangChain](https://github.com/langchain-ai/langchain) 的 Go 语言实现。

## 📖 文档

- [文档站点](https://tmc.github.io/langchaingo/docs/)
- [API 参考](https://pkg.go.dev/github.com/tmc/langchaingo)


## 🎉 示例

请参阅 [./examples](./examples) 获取使用示例。

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
  prompt := "为一家生产彩色袜子的公司取一个好名字会是什么？"
  completion, err := llms.GenerateFromSinglePrompt(ctx, llm, prompt)
  if err != nil {
    log.Fatal(err)
  }
  fmt.Println(completion)
}
```

```shell
$ go run .
七彩袜业 (Socktastic)
```
*(译者注：Socktastic 是一个英文创意词，结合了 Sock (袜子) 和 fantastic (极好的)。这里提供一个中文意译供参考。)*

# 资源

加入 Discord 服务器获取支持和讨论：[加入 Discord](https://discord.gg/8bHGKzHBkM)

以下是一些关于使用 LangChain Go 的博客文章和文章链接：

- [在 Go 中通过 LangChainGo 使用 Gemini 模型](https://eli.thegreenplace.net/2024/using-gemini-models-in-go-with-langchaingo/) - 2024年1月
- [通过 LangChainGo 使用 Ollama](https://eli.thegreenplace.net/2023/using-ollama-with-langchaingo/) - 2023年11月
- [用 Go 创建一个简单的 ChatGPT 克隆](https://sausheong.com/creating-a-simple-chatgpt-clone-with-go-c40b4bec9267?sk=53a2bcf4ce3b0cfae1a4c26897c0deb0) - 2023年8月
- [用 Go 创建一个可以在你的笔记本电脑上运行的 ChatGPT 克隆](https://sausheong.com/creating-a-chatgpt-clone-that-runs-on-your-laptop-with-go-bf9d41f1cf88?sk=05dc67b60fdac6effb1aca84dd2d654e) - 2023年8月


# 贡献者

langchaingo 的发展正趋向于更社区化的努力，如果您有兴趣成为维护者或者您是一位贡献者，请加入我们的 [Discord](https://discord.gg/8bHGKzHBkM) 并告知我们。

<a href="https://github.com/tmc/langchaingo/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=tmc/langchaingo" />
</a>
