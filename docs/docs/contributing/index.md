# 为 LangChainGo 做贡献

感谢您对为 LangChainGo 做贡献的兴趣！本指南帮助您开始。

## 贡献方式

### 1. 代码贡献

- **错误修复**：帮助我们消除错误并提高稳定性
- **新功能**：实现新的 LLM 提供商、工具或链
- **性能改进**：优化现有代码
- **测试**：提高测试覆盖率并添加缺失的测试

### 2. 文档贡献

- **教程**：为常见用例编写分步指南
- **操作指南**：为特定问题创建实用解决方案
- **API 文档**：改进代码注释和示例
- **概念指南**：解释架构决策和模式

### 3. 社区支持

- **回答问题**：在 GitHub 讨论中帮助他人
- **报告问题**：提交详细的错误报告
- **审查 PR**：对拉取请求提供反馈
- **分享示例**：展示您的 LangChainGo 项目

## 开始

### 开发环境设置

1. 在 GitHub 上分叉代码仓库
2. 在本地克隆您的分叉：
   ```bash
   git clone https://github.com/YOUR-USERNAME/langchaingo.git
   cd langchaingo
   ```

3. 添加上游远程：
   ```bash
   git remote add upstream https://github.com/tmc/langchaingo.git
   ```

4. 创建功能分支：
   ```bash
   git checkout -b feature/your-feature-name
   ```

### 代码风格

- 遵循标准的 [Go 约定和惯用法](https://go.dev/doc/effective_go)
- 在提交前运行 `go fmt`
- 确保所有测试通过 `go test ./...`
- 为新功能添加测试
- 使用包前缀的提交消息（见下面的 PR 指南）
- 保持提交专注于单一主题

### 测试

在贡献与外部 API 交互的代码时：

1. 使用内部 `httprr` 工具记录 HTTP 交互
2. 永远不要提交真实的 API 密钥或机密
3. 确保测试可以在没有外部依赖的情况下运行
4. 详细信息请参阅[架构指南](/docs/concepts/architecture#使用-httprr-进行-http-测试)

## 贡献流程

1. **检查现有问题**：查找关于您想法的现有问题或讨论
2. **打开问题**：对于重大更改，先打开问题进行讨论
3. **进行更改**：在功能分支中实现您的更改
4. **遵循提交风格**：使用 Go 风格的包前缀提交消息
5. **彻底测试**：确保所有测试通过并根据需要添加新测试
6. **提交 PR**：按照我们的指南打开带有清晰描述的拉取请求
7. **处理反馈**：及时回应审查意见

## 拉取请求指南

### PR 标题格式

**使用 Go 风格的包前缀提交消息**，遵循 [Go 贡献指南](https://go.dev/doc/contribute#commit_messages)：

- `memory: add interfaces for custom storage backends`
- `llms/openai: fix streaming response handling`
- `chains: implement conversation chain with memory`
- `vectorstores/chroma: add support for metadata filtering`
- `docs: update getting started guide for new API`
- `agents: add tool calling support for GPT-4`
- `examples: add RAG implementation tutorial`

**格式**：`package: description in lowercase without period`

好的提交消息示例：
- `llms/anthropic: implement function calling support`
- `memory: fix buffer overflow in conversation memory`
- `tools: add calculator tool with error handling`
- `all: update dependencies and organize go.mod file`

### PR 描述
包括：
- 更改摘要
- 相关问题编号
- 执行的测试
- 破坏性更改（如有）
- 对 Python/TypeScript LangChain 中类似功能的引用（如适用）

## 文档贡献

查看我们专门的[文档贡献指南](./documentation)了解详细信息：
- 编写教程
- 创建操作指南
- 文档风格指南
- 本地构建和测试文档

## 行为准则

请注意，此项目遵循行为准则。通过参与，您需要遵守此准则。请向项目维护者报告不可接受的行为。

## 认可

贡献者在以下方面得到认可：
- 项目的贡献者列表
- 重大贡献的发布说明
- 书面内容的文档致谢

## 问题？

- 打开 [GitHub 讨论](https://github.com/tmc/langchaingo/discussions)
- 检查现有问题和 PR
- 查看文档

感谢您帮助让 LangChainGo 变得更好！