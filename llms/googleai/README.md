此目录包含针对 Google 模型的 langchaingo 提供程序。

* 在主 `googleai` 目录中：针对 Google AI (https://ai.google.dev/) 的提供程序
* 在 `vertex` 目录中：针对 GCP Vertex AI (https://cloud.google.com/vertex-ai/) 的提供程序
* 在 `palm` 目录中：针对旧版 PaLM 模型的提供程序。

`googleai` 和 `vertex` 提供程序均可访问 Gemini 系列多模态大型语言模型 (LLM)。这些提供程序之间的代码非常相似；因此，`vertex` 包的大部分代码是使用工具从 `googleai` 包代码生成的：

    go run ./llms/googleai/internal/cmd/generate-vertex.go < llms/googleai/googleai.go > llms/googleai/vertex/vertex.go

----

测试：

`googleai` 和 `vertex` 之间的测试代码也是共享的，位于 `shared_test` 目录中。相同的测试会针对这两个提供程序运行。
