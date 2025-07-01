# Chroma 支持

您可以通过包含的 [`vectorstores.VectorStore` 接口](../vectorstores.go) 实现来访问 Chroma，方法是使用 [`New` 函数](./chroma.go) API 创建并使用 Chroma 客户端 `Store` 实例。

## 客户端/服务器模式

在[“内存中”版本](https://docs.trychroma.com/usage-guide#running-chroma-in-clientserver-mode)发布之前，仅提供客户端/服务器模式。

> **注意：** 更多在本地运行 Chroma 的方法可以在 [Chroma 指南](https://cookbook.chromadb.dev/running/running-chroma/)中找到。

创建客户端实例时，使用 [`WithChromaURL` API](./options.go) 或 `CHROMA_URL` 环境变量来指定 Chroma 服务器的 URL。

## 使用 OpenAI LLM

要将 OpenAI LLM 与 Chroma 一起使用，请在创建客户端时使用 [`WithOpenAIAPIKey` API](./options.go) 或 `OPENAI_API_KEY` 环境变量。

## 使用 Docker 运行

在本地 Docker 实例中运行 Chroma 服务器对于测试和开发工作流特别有用。下面提供了一个示例调用场景：

### 启动 Chroma 服务器

在撰写本文时，Chroma Docker 镜像的最新版本是 [chroma:0.5.0](https://github.com/chroma-core/chroma/pkgs/container/chroma/184319417?tag=0.5.0)。
可以通过以下命令直接运行它，同时将其端口暴露给您的本地计算机：

```shell
$ docker run -p 8000:8000 ghcr.io/chroma-core/chroma:0.5.0
```

### 运行示例 `langchaingo` 应用程序

在“简单 Docker 服务器”运行的情况下（见上文），运行包含的示例 `langchaingo` 应用程序应产生以下结果：

```shell
$ export CHROMA_URL=http://localhost:8000
$ export OPENAI_API_KEY=此处填写您的OpenApiKey
$ go run ./examples/chroma-vectorstore-example/chroma_vectorstore_example.go
结果:
1. 案例: 日本最多五个城市
    结果: 东京, 名古屋, 京都, 福冈, 广岛
2. 案例: 南美洲的一个城市
    结果: 布宜诺斯艾利斯
3. 案例: 南美洲的大城市
    结果: 圣保罗, 里约热内卢
```

## 测试

测试套件 `chroma_test.go` 最初是相邻的 `pinecone_test.go` 的克隆版本，内容相对较少。欢迎贡献新的测试用例，或为代码更改添加覆盖范围。
