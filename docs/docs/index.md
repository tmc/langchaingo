# 欢迎来到 LangChainGo

LangChainGo 是 [Go 编程语言](https://go.dev/)版本的 [LangChain](https://www.langchain.com/) (移植/分支)。

LangChain 是一个用于开发由语言模型驱动的应用程序的框架。我们相信，最强大和差异化的应用程序不仅会通过 API 调用语言模型，还会：

- _数据感知 (Be data-aware)_：将语言模型连接到其他数据源
- _具有代理性 (Be agentic)_：允许语言模型与其环境交互

LangChain 框架的设计充分考虑了上述原则。

## 文档结构

_**注意**：这些文档适用于 [LangChainGo](https://github.com/tmc/langchaingo)。_

我们的文档遵循结构化的方法，以帮助您有效地学习和使用 LangChainGo：

### 📚 [教程](./tutorials/)
构建完整应用程序的分步指南。非常适合从头开始学习 LangChainGo。

- **快速入门**：[使用 Ollama 快速设置](./getting-started/guide-ollama.mdx) • [使用 OpenAI 快速设置](./getting-started/guide-openai.mdx)
- **基础应用**：简单的聊天应用、问答系统、文档摘要
- **高级应用**：RAG 系统、带工具的代理、多模态应用
- **生产环境**：部署、优化、监控

### 🛠️ [操作指南](./how-to/)
针对特定问题的实用解决方案。查找 “我该如何...?” 这类问题的答案。

- **LLM 集成**：配置提供商、处理速率限制、实现流式传输
- **文档处理**：加载文档、实现搜索、优化检索
- **代理开发**：创建自定义工具、多步推理、错误处理
- **生产环境**：项目结构、日志记录、部署、扩展

### 🧠 [概念](./concepts/)
对 LangChainGo 架构和设计原则的深入解释。

- **核心架构**：框架设计、接口、Go 特定模式
- **语言模型**：模型抽象、通信模式、优化
- **代理与记忆**：代理模式、记忆管理、状态持久化
- **生产环境**：性能、可靠性、安全注意事项

### 🔧 组件
所有 LangChainGo 模块及其功能的技术参考。

- **模型 I/O (Model I/O)**：LLM、聊天模型、嵌入、提示
- **数据连接 (Data Connection)**：文档加载器、向量存储、文本分割器、检索器
- **[链 (Chains)](./modules/chains/)**：调用序列和端到端应用程序
- **[记忆 (Memory)](./modules/memory/)**：状态持久化和对话管理
- **[代理 (Agents)](./modules/agents/)**：决策制定和自主行为

## API 参考

您可以在[此处](https://pkg.go.dev/github.com/tmc/langchaingo)找到 LangChain 中所有模块的 API 参考，以及所有导出类和函数的完整文档。

## 参与其中

- **[贡献指南](/docs/contributing)**：了解如何贡献代码和文档
- **[GitHub 讨论](https://github.com/tmc/langchaingo/discussions)**：加入关于 LangChainGo 的讨论
- **[GitHub Issues](https://github.com/tmc/langchaingo/issues)**：报告错误或请求功能
