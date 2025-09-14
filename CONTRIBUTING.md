# 为 langchaingo 做贡献

首先，感谢您抽出时间做出贡献！ ❤️

我们鼓励并重视所有类型的贡献。请参阅[目录](#目录)了解不同的帮助方式以及本项目如何处理它们的详细信息。在做出贡献之前，请务必阅读相关部分。这将使我们维护者的工作更加轻松，并使所有相关人员的体验更加顺畅。社区期待您的贡献。 🎉

> 如果您喜欢这个项目，但没有时间做出贡献，那也没关系。还有其他简单的方法可以支持项目并表达您的谢意，我们对此也会非常高兴：
> - 为项目点赞 (Star)
> - 在社交媒体上宣传它 (Tweet about it)
> - 在您项目的自述文件中引用此项目
> - 在本地聚会上提及该项目并告诉您的朋友/同事

## 目录

- [行为准则](#行为准则)
- [我有一个问题](#我有一个问题)
- [我想做出贡献](#我想做出贡献)
  - [报告错误](#报告错误)
    - [提交错误报告之前](#提交错误报告之前)
    - [如何提交一个好的错误报告？](#如何提交一个好的错误报告)
  - [建议增强功能](#建议增强功能)
    - [提交增强功能建议之前](#提交增强功能建议之前)
    - [如何提交一个好的增强功能建议？](#如何提交一个好的增强功能建议)
  - [您的第一个代码贡献](#您的第一个代码贡献)
    - [进行更改](#进行更改)
      - [在用户界面中进行更改](#在用户界面中进行更改)
      - [在本地进行更改](#在本地进行更改)
    - [运行测试](#运行测试)
    - [使用 httprr 进行测试](#使用-httprr-进行测试)
      - [httprr 如何工作](#httprr-如何工作)
      - [使用 httprr 编写测试](#使用-httprr-编写测试)
      - [录制新测试](#录制新测试)
      - [关于 httprr 的重要说明](#关于-httprr-的重要说明)
      - [调试 httprr 问题](#调试-httprr-问题)
    - [提交您的更新](#提交您的更新)
    - [拉取请求 (Pull Request)](#拉取请求-pull-request)
    - [您的 PR 已合并！](#您的-pr-已合并)

## 行为准则

本项目及其所有参与者均受 [langchaingo 行为准则](CODE_OF_CONDUCT.md)的约束。
通过参与，您应遵守此准则。请将不可接受的行为报告给 <travis.cline@gmail.com>。


## 我有一个问题

> 如果您想提问，我们假设您已经阅读了可用的[文档](https://pkg.go.dev/github.com/tmc/langchaingo)。

在提问之前，最好先搜索现有的[问题 (Issues)](https://github.com/tmc/langchaingo/issues) 看是否能帮到您。如果您找到了合适的问题但仍需要澄清，您可以在该问题中写下您的问题。建议您也先在互联网上搜索答案。

如果您仍然觉得需要提问并需要澄清，我们建议如下：

- 新建一个[问题 (Issue)](https://github.com/tmc/langchaingo/issues/new)。
- 尽可能多地提供您遇到的问题的上下文。
- 提供项目和平台版本（nodejs、npm 等），具体取决于哪些信息似乎相关。

然后我们会尽快处理该问题。

## 我想做出贡献

> ### 法律声明
> 当您为本项目做出贡献时，您必须同意您是所贡献内容的 100% 作者，您拥有对该内容的必要权利，并且您贡献的内容可以在项目许可下提供。

### 报告错误

#### 提交错误报告之前

一个好的错误报告不应该让其他人需要向您追问更多信息。因此，我们要求您仔细调查、收集信息并在报告中详细描述问题。请提前完成以下步骤，以帮助我们尽快修复任何潜在的错误。

- 确保您使用的是最新版本。
- 判断您的错误是否真的是一个错误，而不是您操作不当造成的，例如使用了不兼容的环境组件/版本（请确保您已阅读[文档](https://pkg.go.dev/github.com/tmc/langchaingo)。如果您正在寻求支持，您可能需要查看[此部分](#我有一个问题)）。
- 查看其他用户是否遇到过（并可能已经解决了）与您相同的问题，请在[错误跟踪器](https://github.com/tmc/langchaingo/issues?q=label%3Abug)中检查是否已存在针对您的错误或问题的错误报告。
- 同时确保搜索互联网（包括 Stack Overflow），看看 GitHub 社区之外的用户是否讨论过该问题。
- 收集有关错误的信息：
  - 堆栈跟踪 (Stack trace / Traceback)
  - 操作系统、平台和版本 (Windows、Linux、macOS、x86、ARM)
  - 解释器、编译器、SDK、运行时环境、包管理器的版本，具体取决于哪些信息似乎相关。
  - 可能的话，提供您的输入和输出
  - 您能可靠地重现该问题吗？您也能用旧版本重现它吗？

#### 如何提交一个好的错误报告？

> 您绝不能将与安全相关的问题、漏洞或包含敏感信息的错误报告给问题跟踪器或其他公共场所。相反，敏感错误必须通过电子邮件发送至 <travis.cline@gmail.com>。
<!-- 您可以添加一个 PGP 密钥以允许加密发送消息。 -->

我们使用 GitHub Issues 来跟踪错误。如果您遇到项目问题：

- 新建一个[问题 (Issue)](https://github.com/tmc/langchaingo/issues/new)。（由于此时我们不确定这是否是一个错误，我们要求您暂时不要称其为错误，也不要给问题添加标签。）
- 解释您期望的行为和实际行为。
- 请尽可能多地提供上下文，并描述其他人可以遵循的*重现步骤*，以便他们自己重现该问题。这通常包括您的代码。对于好的错误报告，您应该隔离问题并创建一个简化的测试用例。
- 提供您在上一节中收集的信息。

提交后：

- 项目团队将相应地标记问题。
- 团队成员将尝试使用您提供的步骤重现该问题。如果没有重现步骤或没有明显的方法可以重现问题，团队将要求您提供这些步骤，并将问题标记为 `needs-repro`。带有 `needs-repro` 标签的错误在重现之前不会得到处理。
- 如果团队能够重现该问题，它将被标记为 `needs-fix`，以及可能的其他标签（例如 `critical`），并且该问题将留给[某人来实现](#您的第一个代码贡献)。

<!-- 您可能希望为错误创建一个问题模板，该模板可用作指南并定义要包含的信息的结构。如果这样做，请在此处的描述中引用它。 -->


### 建议增强功能

本节指导您为 langchaingo 提交增强功能建议，**包括全新的功能和对现有功能的微小改进**。遵循这些准则将有助于维护者和社区理解您的建议并找到相关的建议。

#### 提交增强功能建议之前

- 确保您使用的是最新版本。
- 仔细阅读[文档](https://pkg.go.dev/github.com/tmc/langchaingo)并确定该功能是否已被涵盖，也许可以通过单独的配置实现。
- 执行[搜索](https://github.com/tmc/langchaingo/issues)以查看是否已有人建议过此增强功能。如果是，请向现有问题添加评论，而不是新建一个。
- 确定您的想法是否符合项目的范围和目标。您需要充分说明理由，以说服项目的开发人员相信此功能的优点。请记住，我们希望功能对我们的大多数用户有用，而不仅仅是一小部分用户。如果您只针对少数用户，请考虑编写一个附加组件/插件库。

#### 如何提交一个好的增强功能建议？

增强功能建议作为 [GitHub Issues](https://github.com/tmc/langchaingo/issues) 进行跟踪。

- 为问题使用**清晰且描述性的标题**以识别建议。
- 尽可能详细地**逐步描述建议的增强功能**。
- **描述当前行为**并**解释您期望看到的行为以及原因**。此时，您还可以说明哪些替代方案对您不起作用。
- 您可能希望**包含屏幕截图和动画 GIF**，以帮助您演示步骤或指出建议相关的部分。您可以使用[此工具](https://www.cockos.com/licecap/)在 macOS 和 Windows 上录制 GIF，以及在 Linux 上使用[此工具](https://github.com/colinkeenan/silentcast)或[此工具](https://github.com/GNOME/byzanz)。<!-- 仅当项目具有 GUI 时才应包含此内容 -->
- **解释为什么此增强功能对大多数 langchaingo 用户有用**。您可能还想指出其他项目如何更好地解决了该问题，哪些项目可以作为灵感。
- 我们力求在概念上与 Langchain 的 Python 和 TypeScript 版本保持一致。在引入新概念时，请链接/引用这些代码库中的相关概念。

<!-- 您可能希望为增强功能建议创建一个问题模板，该模板可用作指南并定义要包含的信息的结构。如果这样做，请在此处的描述中引用它。 -->

### 您的第一个代码贡献

#### 进行更改

##### 在用户界面中进行更改

单击任何文档页面底部的**做出贡献 (Make a contribution)** 以进行小的更改，例如拼写错误、句子修正或损坏的链接。这将带您进入 `.md` 文件，您可以在其中进行更改并[创建拉取请求](#拉取请求-pull-request)以供审阅。

##### 在本地进行更改

1. Fork (分叉) 代码仓库。
- 使用 GitHub Desktop：
  - [GitHub Desktop 入门](https://docs.github.com/en/desktop/installing-and-configuring-github-desktop/getting-started-with-github-desktop)将指导您完成 Desktop 的设置。
  - 设置好 Desktop 后，您可以使用它来 [fork 代码仓库](https://docs.github.com/en/desktop/contributing-and-collaborating-using-github-desktop/cloning-and-forking-repositories-from-github-desktop)！

- 使用命令行：
  - [Fork 代码仓库](https://docs.github.com/en/github/getting-started-with-github/fork-a-repo#fork-an-example-repository)，以便您可以在不影响原始项目的情况下进行更改，直到准备好合并它们。

2. 安装或确保 **Golang** 已更新。

3. 创建一个工作分支并开始进行更改！

##### 最近的更新和依赖项

贡献时请注意这些最近的更改：

- **HTTP 客户端标准化**：所有 HTTP 客户端现在都使用带有自定义 User-Agent 标头 (`langchaingo/{version}`) 的 `httputil.DefaultClient`
- **HuggingFace 环境变量**：按优先级顺序支持多个令牌来源：`HF_TOKEN`、`HUGGINGFACEHUB_API_TOKEN`、来自 `HF_TOKEN_PATH` 的令牌文件，或默认的 `~/.cache/huggingface/token`
- **OpenAI Functions Agent**：已更新以处理 OpenAI 新的工具调用 API，同时保持向后兼容性
- **Chroma 向量存储**：已更新为使用 `github.com/amikos-tech/chroma-go` v0.1.4+
- **Testcontainers 迁移**：在支持的情况下，新的 testcontainers API 使用 `Run()` 而不是已弃用的 `RunContainer()`
- **HTTPRR 文件**：不再压缩 - 直接将 `.httprr` 文件提交到代码仓库

##### 项目结构和约定

进行更改时，请遵循以下架构约定：

- **HTTP 客户端**：对所有 HTTP 操作使用 `httputil.DefaultClient` 而不是 `http.DefaultClient`，以确保正确的 User-Agent 标头
- **基于接口的设计**：核心功能通过接口（Model、Chain、Memory 等）定义
- **提供程序隔离**：每个 LLM/嵌入提供程序都有其自己的包，其中包含内部客户端实现
- **选项模式 (Options Pattern)**：对配置使用功能选项（请参阅现有示例）
- **上下文传播 (Context Propagation)**：所有操作都应接受 `context.Context` 以进行取消和设置截止日期
- **错误处理**：使用标准化的错误类型和映射（请参阅 `llms.Error` 和提供程序错误映射器）

##### 添加新的 LLM 提供程序

添加新的 LLM 提供程序时：

1. 在 `/llms/your-provider` 下创建一个新包
2. 实现 `llms.Model` 接口
3. 为 HTTP 交互创建一个内部客户端包
4. 对 HTTP 请求使用 `httputil.DefaultClient`
5. 添加合规性测试：`compliance.NewSuite("yourprovider", model).Run(t)`
6. 为 HTTP 调用添加带有 httprr 录制的测试
7. 遵循现有的提供程序模式进行选项和错误处理

##### 添加新的向量存储

添加新的向量存储时：

1. 在 `/vectorstores/your-store` 下创建一个新包
2. 实现向量存储接口
3. 尽可能使用 testcontainers 进行集成测试
4. 遵循现有的距离策略和元数据过滤模式

#### 运行测试

提交更改之前，请确保所有测试都通过：

```bash
# 运行所有测试
make test

# 运行特定包的测试
go test ./chains

# 运行特定测试
go test -run TestLLMChain ./chains

# 运行带有竞争检测的测试
make test-race

# 运行带有覆盖率的测试
make test-cover

# 测试分离脚本
./scripts/run_unit_tests.sh      # 仅运行单元测试（无外部依赖）
./scripts/run_all_tests.sh       # 运行完整的测试套件
./scripts/run_integration_tests.sh # 仅运行集成测试（需要 Docker）

# 录制测试的 HTTP 交互（添加新测试时）
go test -httprecord=. -v ./path/to/package
```

同时确保您的代码通过 linting (代码风格检查)：

```bash
# 运行 linter
make lint

# 运行 linter 并自动修复
make lint-fix

# 运行实验性 linter 配置
make lint-exp

# 运行所有 linter，包括实验性的
make lint-all

# 清理 lint 缓存
make clean-lint-cache

# 开发工具
make build-examples         # 构建所有示例以验证它们是否可以编译
make docs                  # 生成文档
make run-pkgsite          # 运行本地文档服务器
make install-git-hooks    # 安装 git钩子 (设置 pre-push 钩子)
make pre-push             # 运行 lint 和快速测试 (适用于 git pre-push 钩子)
```

##### 其他开发工具

项目在 `/internal/devtools` 中包含多个开发工具：

```bash
# 自定义 linting 工具
make lint-devtools         # 运行自定义架构 lint
make lint-devtools-fix     # 运行自定义 lint 并自动修复
make lint-architecture     # 运行架构验证
make lint-prepush          # 运行 pre-push lint
make lint-prepush-fix      # 运行 pre-push lint 并自动修复

# HTTPRR 管理
go run ./internal/devtools/rrtool list-packages  # 列出使用 httprr 的包
make test-record           # 重新录制所有 HTTP 交互

# 测试模式验证
make lint-testing          # 检查不正确的 httprr 测试模式
make lint-testing-fix      # 尝试自动修复 httprr 测试模式
```

#### 使用 httprr 进行测试

本项目使用自定义的 HTTP 录制/回放系统 (httprr) 来测试与外部 API 的 HTTP 交互。这使得测试可以确定性地运行，而无需实际的 API 凭据或进行真实的 API 调用。

##### httprr 如何工作

- **录制模式**：当测试使用真实的 API 凭据运行时，httprr 会将所有 HTTP 请求和响应录制到 `testdata` 目录中的 `.httprr` 文件中。
- **回放模式**：当测试在没有凭据的情况下运行时，httprr 会从 `.httprr` 文件中回放录制的 HTTP 交互。
- **自动模式切换**：如果没有可用的凭据且没有录制文件，测试会自动跳过，并显示有用的消息。

##### 使用 httprr 编写测试

在编写调用外部 API 的 HTTP 测试时，请遵循以下模式：

```go
func TestMyFeature(t *testing.T) {
    t.Parallel()
    ctx := context.Background()

    // 如果没有凭据且缺少录制文件，则跳过
    httprr.SkipIfNoCredentialsAndRecordingMissing(t, "OPENAI_API_KEY")

    // 设置 httprr (通过 t.Cleanup 自动清理)
    // 对 User-Agent 标头使用 httputil.DefaultTransport，或对更简单的情况使用 http.DefaultTransport
    rr := httprr.OpenForTest(t, httputil.DefaultTransport)

    var opts []openai.Option
    opts = append(opts, openai.WithHTTPClient(rr.Client()))

    // 回放时使用测试令牌
    if !rr.Recording() {
        opts = append(opts, openai.WithToken("test-api-key"))
    }
    // 录制时，客户端将使用环境中的真实 API 密钥

    client, err := openai.New(opts...)
    require.NoError(t, err)

    // 运行您的测试
    result, err := client.Call(ctx, "test input")
    require.NoError(t, err)
    // ...断言...
}
```

此模式可确保：
- **录制时**：使用环境中的真实 API 密钥来捕获有效的响应
- **回放时**：使用 "test-api-key" 来满足客户端验证（httprr 在实际 API 调用之前进行拦截）

对于其他提供程序，请使用其特定选项：

```go
// HuggingFace 示例 (支持多个环境变量)
func TestHuggingFace(t *testing.T) {
    // HuggingFace 支持 HF_TOKEN 和 HUGGINGFACEHUB_API_TOKEN
    if os.Getenv("HF_TOKEN") == "" && os.Getenv("HUGGINGFACEHUB_API_TOKEN") == "" {
        httprr.SkipIfNoCredentialsAndRecordingMissing(t, "HF_TOKEN")
    }

    rr := httprr.OpenForTest(t, httputil.DefaultTransport)

    apiKey := "test-api-key"
    if rr.Recording() {
        if key := os.Getenv("HF_TOKEN"); key != "" {
            apiKey = key
        } else if key := os.Getenv("HUGGINGFACEHUB_API_TOKEN"); key != "" {
            apiKey = key
        }
    }

    llm, err := huggingface.New(
        huggingface.WithHTTPClient(rr.Client()),
        huggingface.WithToken(apiKey),
    )
    // ...
}

// Perplexity 示例
var opts []perplexity.Option
opts = append(opts, perplexity.WithHTTPClient(rr.Client()))
if !rr.Recording() {
    opts = append(opts, perplexity.WithAPIKey("test-api-key"))
}
tool, err := perplexity.New(opts...)

// SerpAPI 示例，带有请求清理功能
rr.ScrubReq(func(req *http.Request) error {
    if req.URL != nil {
        q := req.URL.Query()
        q.Set("api_key", "test-api-key")
        req.URL.RawQuery = q.Encode()
    }
    return nil
})
```

对于需要多次创建客户端的测试，请考虑使用辅助函数：

```go
func newOpenAILLM(t *testing.T) *openai.LLM {
    t.Helper()
    httprr.SkipIfNoCredentialsAndRecordingMissing(t, "OPENAI_API_KEY")

    rr := httprr.OpenForTest(t, httputil.DefaultTransport)

    // 仅在不录制时并行运行测试（以避免速率限制）
    if !rr.Recording() {
        t.Parallel()
    }

    var opts []openai.Option
    opts = append(opts, openai.WithHTTPClient(rr.Client()))

    if !rr.Recording() {
        opts = append(opts, openai.WithToken("test-api-key"))
    }
    // 录制时，openai.New() 将从环境中读取 OPENAI_API_KEY

    llm, err := openai.New(opts...)
    require.NoError(t, err)
    return llm
}
```

##### 录制新测试

为新测试录制 HTTP 交互：

1. 设置所需的环境变量（例如 `OPENAI_API_KEY`）
2. 启用录制并运行测试：
   ```bash
   go test -v -httprecord=. ./path/to/package

   # 为避免速率限制，可以控制并行度：
   go test -v -httprecord=. -p 1 -parallel=1 ./path/to/package

   # 或者使用 Makefile 目标来录制所有包
   make test-record
   ```
3. 测试将在 `testdata` 目录中创建 `.httprr` 文件
4. 将这些录制文件与您的 PR 一起提交
5. 对于需要 API 密钥清理的测试，请添加请求清理函数

##### 关于 httprr 的重要说明

- **Transport 选择**：对 User-Agent 标头使用 `httputil.DefaultTransport`，或对更简单的情况使用 `http.DefaultTransport`
- **检查 rr.Recording()**：使用此选项仅在回放时有条件地添加测试令牌
- **httprr 处理清理**：OpenForTest 会自动使用 t.Cleanup() 注册清理
- **录制时使用真实密钥**：录制时，让客户端使用环境中的真实 API 密钥
- **回放时使用测试令牌**：回放时，使用 "test-api-key" 来满足客户端验证
- **并行测试**：仅在不录制时运行 `t.Parallel()` 以避免达到 API 速率限制
- **多个凭据来源**：对于 HuggingFace，同时检查 `HF_TOKEN` 和 `HUGGINGFACEHUB_API_TOKEN`
- **请求清理**：对需要 URL 参数清理的 API（如 SerpAPI）使用 `rr.ScrubReq()`
- **录制是确定性的**：相同的输入应产生相同的输出
- **敏感数据已清理**：httprr 会自动从录制中删除授权标头和其他敏感数据
- **提交录制文件**：始终提交 `.httprr` 文件，以便测试可以在没有凭据的 CI 中运行
- **删除无效录制**：如果由于录制无效（例如 401 错误）导致测试失败，请删除录制文件并使用有效凭据重新录制

##### 调试 httprr 问题

- 使用 `-httprecord-debug` 标志获取详细的录制信息
- 使用 `-httpdebug` 标志查看实际的 HTTP 流量
- 检查录制是否存在：`ls testdata/*.httprr`
- 验证录制内容：`head testdata/TestName.httprr`
- 使用测试分离脚本来隔离单元测试与集成测试问题：
  ```bash
  ./scripts/run_unit_tests.sh      # 无外部依赖的快速测试
  ./scripts/run_integration_tests.sh # 需要 Docker/外部服务的测试
  ```

##### 自动化的 httprr 模式验证

项目包含一个自定义 linter 来检测不正确的 httprr 使用模式：

```bash
# 检查不正确的模式
make lint-testing

# 查看发现的具体问题
go run ./internal/devtools/lint -testing -v
```

linter 会检测：
- **硬编码的测试令牌**：无条件调用 `WithToken("test-api-key")`（应根据 `!rr.Recording()` 条件调用）
- **不正确的并行执行**：在 httprr 设置之前调用 `t.Parallel()`（应根据 `!rr.Recording()` 条件调用）

这些问题会导致录制期间出现身份验证错误以及测试期间出现竞争条件。

#### 提交您的更新

当您对更改满意后，提交它们。不要忘记进行自我审查以加快审查过程：zap:。

#### 拉取请求 (Pull Request)

完成更改后，创建一个拉取请求，也称为 PR。
- 根据 [Go 贡献指南](https://go.dev/doc/contribute#commit_messages)清晰、简洁地命名您的拉取请求标题，并以您更改的主要受影响包的名称作为前缀。（例如 `memory: add interfaces` 或 `util: add helpers`）
- 运行所有 linter 并确保测试通过：`make lint && make test`
- 如果您添加了新的基于 HTTP 的功能，请包含 httprr 录制
- **我们力求在概念上与 Langchain 的 Python 和 TypeScript 版本保持一致。在引入新概念时，请链接/引用这些代码库中的相关概念。**
- 填写“准备审阅”模板，以便我们可以审阅您的 PR。此模板可帮助审阅者了解您的更改以及拉取请求的目的。
- 如果您正在解决某个问题，请不要忘记[将 PR 链接到问题](https://docs.github.com/en/issues/tracking-your-work-with-issues/linking-a-pull-request-to-an-issue)。
- 选中复选框以[允许维护者编辑](https://docs.github.com/en/github/collaborating-with-issues-and-pull-requests/allowing-changes-to-a-pull-request-branch-created-from-a-fork)，以便可以更新分支以进行合并。
提交 PR 后，团队成员将审阅您的提议。我们可能会提出问题或要求提供其他信息。
- 我们可能会要求在合并 PR 之前进行更改，可以使用[建议的更改](https://docs.github.com/en/github/collaborating-with-issues-and-pull-requests/incorporating-feedback-in-your-pull-request)或拉取请求评论。您可以直接通过 UI 应用建议的更改。您可以在您的 fork 中进行任何其他更改，然后将它们提交到您的分支。
- 在更新 PR 并应用更改时，将每个对话标记为[已解决](https://docs.github.com/en/github/collaborating-with-issues-and-pull-requests/commenting-on-a-pull-request#resolving-conversations)。
- 如果遇到任何合并问题，请查看此 [git 教程](https://github.com/skills/resolve-merge-conflicts)以帮助您解决合并冲突和其他问题。

#### 您的 PR 已合并！

恭喜 :tada::tada: langchaingo 团队感谢您 :sparkles:。

PR 合并后，您的贡献将在代码仓库贡献者列表中公开可见。

现在您已成为社区的一员！
