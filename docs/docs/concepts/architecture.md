# LangChainGo 架构

本文档解释了 LangChainGo 的架构以及它如何遵循 Go 语言约定。

## 模块化采用理念

**您无需采用整个 LangChainGo 框架。** 架构设计支持选择性采用 - 只使用解决您特定问题的组件：

- **需要 LLM 客户端？** 只使用 `llms` 包
- **想要提示模板？** 添加 `prompts` 包
- **构建对话应用？** 包含 `memory` 进行状态管理
- **创建自主代理？** 组合 `agents`、`tools` 和 `chains`

每个组件都设计为独立工作，同时在组合时提供无缝集成。从小处开始，根据需要扩展您的使用。

## 标准库对齐

LangChainGo 遵循 Go 标准库的模式和理念。我们的接口模仿经过验证的标准库设计：

- **`context.Context` 优先**：如 `database/sql`、`net/http` 和其他标准库包
- **接口组合**：小型、专注的接口，能够很好地组合（如 `io.Reader`、`io.Writer`）
- **构造器模式**：带有函数选项的 `New()` 函数（如 `http.Client`）
- **错误处理**：带有类型断言的显式错误（如 `net.OpError`、`os.PathError`）

当标准库演进时，我们也随之演进。最近的例子：
- 采用 `slog` 模式进行结构化日志记录
- 使用 `context.WithCancelCause` 进行更丰富的取消操作
- 遵循 `testing/slogtest` 模式进行处理器验证

### 接口演进

我们的核心接口将随着 Go 和 AI 生态系统的发展而变化。我们欢迎关于更好地与标准库模式对齐的讨论 - 如果您看到使我们的 API 更符合 Go 风格的机会，请开启一个问题。

常见的改进领域：
- 与标准库约定一致的方法命名
- 错误类型定义和处理模式
- 匹配 `io` 包设计的流式模式
- 遵循标准库示例的配置模式

## 设计理念

LangChainGo 围绕几个关键原则构建：

### 接口驱动设计

每个主要组件都由接口定义：
- **模块化**：无需更改代码即可交换实现
- **可测试性**：模拟接口进行测试
- **可扩展性**：添加新的提供商和组件

```go
type Model interface {
    GenerateContent(ctx context.Context, messages []MessageContent, options ...CallOption) (*ContentResponse, error)
}

type Chain interface {
    Call(ctx context.Context, inputs map[string]any, options ...ChainCallOption) (map[string]any, error)
    GetMemory() schema.Memory
    GetInputKeys() []string
    GetOutputKeys() []string
}
```

### 上下文优先方法

所有操作都接受 `context.Context` 作为第一个参数：
- **取消**：取消长时间运行的操作
- **超时**：为 API 调用设置截止时间
- **请求追踪**：通过调用栈传播请求上下文
- **优雅关闭**：干净地处理应用程序终止

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

response, err := llm.GenerateContent(ctx, messages)
```

### Go 惯用模式

#### 错误处理
错误处理使用 Go 的标准模式和类型化错误：

```go
type Error struct {
    Code    ErrorCode
    Message string
    Cause   error
}

// 检查特定错误类型
if errors.Is(err, llms.ErrRateLimit) {
    // 处理速率限制
}
```

#### 选项模式
函数选项提供灵活的配置：

```go
llm, err := openai.New(
    openai.WithModel("gpt-4"),
    openai.WithTemperature(0.7),
    openai.WithMaxTokens(1000),
)
```

#### 通道和 goroutine
使用 Go 的并发特性进行流式处理和并行处理：

```go
// 流式响应
response, err := llm.GenerateContent(ctx, messages, 
    llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
        select {
        case resultChan <- chunk:
        case <-ctx.Done():
            return ctx.Err()
        }
        return nil
    }),
)
```

## 核心组件

### 1. 模型层

模型层为不同类型的语言模型提供抽象：

```
┌─────────────────────────────────────────────────────────┐
│                    模型层                               │
├─────────────────┬─────────────────┬─────────────────────┤
│   聊天模型      │   LLM 模型      │    嵌入模型         │
├─────────────────┼─────────────────┼─────────────────────┤
│  • OpenAI       │  • 补全         │  • OpenAI           │
│  • Anthropic    │  • 传统 API     │  • HuggingFace      │
│  • Google AI    │  • 本地模型     │  • 本地模型         │
│  • 本地 (Ollama)│                │                     │
└─────────────────┴─────────────────┴─────────────────────┘
```

每种模型类型实现特定接口：
- `Model`：所有语言模型的统一接口
- `EmbeddingModel`：专门用于生成嵌入
- `ChatModel`：针对对话交互优化

### 2. 提示管理

提示是一等公民，支持模板：

```go
template := prompts.NewPromptTemplate(
    "你是一个 {{.role}}。回答这个问题：{{.question}}",
    []string{"role", "question"},
)

prompt, err := template.Format(map[string]any{
    "role":     "有用的助手",
    "question": "什么是 Go？",
})
```

### 3. 内存子系统

内存提供有状态的对话管理：

```
┌─────────────────────────────────────────────────────────┐
│                  内存子系统                             │
├─────────────────┬─────────────────┬─────────────────────┤
│  缓冲内存       │  窗口内存       │    摘要内存         │
├─────────────────┼─────────────────┼─────────────────────┤
│ • 简单缓冲      │ • 滑动窗口      │ • 自动摘要          │
│ • 完整历史      │ • 固定大小      │ • 令牌管理          │
│ • 快速访问      │ • 关注最近      │ • 长对话            │
└─────────────────┴─────────────────┴─────────────────────┘
```

### 4. 链编排

链支持复杂的工作流：

```go
// 顺序链示例
chain1 := chains.NewLLMChain(llm, template1)
chain2 := chains.NewLLMChain(llm, template2)

// 对于简单的顺序链，其中一个的输出作为下一个的输入
sequential := chains.NewSimpleSequentialChain([]chains.Chain{chain1, chain2})

// 或者对于具有特定输入/输出键的复杂顺序链
sequential, err := chains.NewSequentialChain(
    []chains.Chain{chain1, chain2},
    []string{"input"},        // 输入键
    []string{"final_output"}, // 输出键
)
```

### 5. 代理框架

代理提供自主行为：

```
┌─────────────────────────────────────────────────────────┐
│                  代理框架                               │
├─────────────────┬─────────────────┬─────────────────────┤
│     代理        │     工具        │    执行器           │
├─────────────────┼─────────────────┼─────────────────────┤
│ • 决策逻辑      │ • 计算器        │ • 执行循环          │
│ • 工具选择      │ • 网络搜索      │ • 错误处理          │
│ • ReAct 模式    │ • 文件操作      │ • 结果处理          │
│ • 规划          │ • 自定义工具    │ • 内存管理          │
└─────────────────┴─────────────────┴─────────────────────┘
```

## 数据流

### 请求流
```
用户输入 → 提示模板 → LLM → 输出解析器 → 响应
     ↓          ↓         ↓        ↓        ↓
   内存 ←── 链逻辑 ←── API 调用 ←── 处理 ←── 内存
```

### 代理流
```
用户输入 → 代理规划 → 工具选择 → 工具执行
     ↓          ↓          ↓          ↓
   内存 ←── 结果分析 ←── 工具结果 ←── 外部 API
     ↓          ↓
   响应 ←── 最终答案
```

## 并发模型

LangChainGo 拥抱 Go 的并发模型：

### 并行处理
```go
// 并发处理多个输入
var wg sync.WaitGroup
results := make(chan string, len(inputs))

for _, input := range inputs {
    wg.Add(1)
    go func(inp string) {
        defer wg.Done()
        result, err := chain.Run(ctx, inp)
        if err == nil {
            results <- result
        }
    }(input)
}

wg.Wait()
close(results)
```

### 流式处理
```go
// 使用通道进行流式处理
type StreamProcessor struct {
    input  chan string
    output chan string
}

func (s *StreamProcessor) Process(ctx context.Context) {
    for {
        select {
        case input := <-s.input:
            // 处理输入
            result := processInput(input)
            s.output <- result
        case <-ctx.Done():
            return
        }
    }
}
```

## 扩展点

### 自定义 LLM 提供商
实现 `Model` 接口：

```go
type CustomLLM struct {
    apiKey string
    client *http.Client
}

func (c *CustomLLM) GenerateContent(ctx context.Context, messages []MessageContent, options ...CallOption) (*ContentResponse, error) {
    // 自定义实现
}
```

### 自定义工具
实现 `Tool` 接口：

```go
type CustomTool struct {
    name        string
    description string
}

func (t *CustomTool) Name() string { return t.name }
func (t *CustomTool) Description() string { return t.description }
func (t *CustomTool) Call(ctx context.Context, input string) (string, error) {
    // 工具逻辑
}
```

### 自定义内存
实现 `Memory` 接口：

```go
type CustomMemory struct {
    storage map[string][]MessageContent
}

func (m *CustomMemory) ChatHistory() schema.ChatMessageHistory {
    // 返回聊天历史实现
}

func (m *CustomMemory) MemoryVariables() []string {
    return []string{"history"}
}
```

## 性能考虑

### 连接池
LLM 提供商使用 HTTP 连接池来提高效率：

```go
client := &http.Client{
    Transport: &http.Transport{
        MaxIdleConns:        100,
        MaxIdleConnsPerHost: 10,
        IdleConnTimeout:     90 * time.Second,
    },
}
```

### 内存管理
- 为您的用例使用适当的内存类型
- 为长时间运行的应用程序实现清理策略
- 在生产环境中监控内存使用

### 缓存
在多个层级实现缓存：
- LLM 响应缓存
- 嵌入缓存
- 工具结果缓存

```go
type CachingLLM struct {
    llm   Model
    cache map[string]*ContentResponse
    mutex sync.RWMutex
}
```

## 错误处理策略

### 分层错误处理
1. **提供商层级**：处理 API 特定错误
2. **组件层级**：处理组件特定错误
3. **应用层级**：处理业务逻辑错误

### 重试逻辑
```go
func retryableCall(ctx context.Context, fn func() error) error {
    backoff := time.Second
    maxRetries := 3
    
    for i := 0; i < maxRetries; i++ {
        err := fn()
        if err == nil {
            return nil
        }
        
        if !isRetryable(err) {
            return err
        }
        
        select {
        case <-time.After(backoff):
            backoff *= 2
        case <-ctx.Done():
            return ctx.Err()
        }
    }
    
    return fmt.Errorf("超过最大重试次数")
}
```

## 测试架构

### 接口模拟
使用接口进行全面测试：

```go
type MockLLM struct {
    responses []string
    index     int
}

func (m *MockLLM) GenerateContent(ctx context.Context, messages []MessageContent, options ...CallOption) (*ContentResponse, error) {
    if m.index >= len(m.responses) {
        return nil, fmt.Errorf("没有更多响应")
    }
    
    response := &ContentResponse{
        Choices: []ContentChoice{{Content: m.responses[m.index]}},
    }
    m.index++
    return response, nil
}
```

### 使用 httprr 进行 HTTP 测试

对于基于 HTTP 的 LLM 提供商的内部测试，LangChainGo 使用 [httprr](https://pkg.go.dev/github.com/tmc/langchaingo/internal/httprr) 来记录和重放 HTTP 交互。这是 LangChainGo 自己的测试套件使用的内部测试工具，确保可靠、快速的测试，而无需访问真实的 API。

#### 设置 httprr

```go
func TestOpenAIWithRecording(t *testing.T) {
    // 启动 httprr 记录器
    recorder := httprr.New("testdata/openai_recording")
    defer recorder.Stop()
    
    // 配置 HTTP 客户端使用记录器
    client := &http.Client{
        Transport: recorder,
    }
    
    // 创建带有自定义客户端的 LLM
    llm, err := openai.New(
        openai.WithHTTPClient(client),
        openai.WithToken("test-token"), // 将在记录中被编辑
    )
    require.NoError(t, err)
    
    // 进行实际 API 调用（首次运行时记录，后续运行时重放）
    response, err := llm.GenerateContent(context.Background(), []llms.MessageContent{
        llms.TextParts(llms.ChatMessageTypeHuman, "你好，世界！"),
    })
    require.NoError(t, err)
    require.NotEmpty(t, response.Choices[0].Content)
}
```

#### 记录指南

1. **初始记录**：使用真实的 API 凭据运行测试以创建记录
2. **敏感数据**：httprr 自动编辑常见的敏感标头
3. **确定性测试**：记录确保跨环境的一致测试结果
4. **版本控制**：提交记录文件以保持团队一致性

#### 使用 httprr 贡献

在为 LangChainGo 的内部测试做贡献时：

1. **为新的 LLM 提供商使用 httprr**：
   ```go
   func TestNewProvider(t *testing.T) {
       recorder := httprr.New("testdata/newprovider_test")
       defer recorder.Stop()
       
       // 测试实现
   }
   ```

2. **当 API 更改时更新记录**：
   ```bash
   # 删除旧记录
   rm testdata/provider_test.httprr
   
   # 使用真实凭据重新运行测试
   PROVIDER_API_KEY=real_key go test
   ```

3. **验证记录已提交**：
   ```bash
   git add testdata/*.httprr
   git commit -m "test: 更新 API 记录"
   ```

### 集成测试
使用 testcontainers 处理外部依赖：

```go
func TestWithDatabase(t *testing.T) {
    ctx := context.Background()
    
    // 使用 testcontainers 启动数据库
    container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
        ContainerRequest: testcontainers.ContainerRequest{
            Image:        "postgres:13",
            ExposedPorts: []string{"5432/tcp"},
            Env: map[string]string{
                "POSTGRES_PASSWORD": "password",
                "POSTGRES_DB":       "testdb",
            },
        },
        Started: true,
    })
    require.NoError(t, err)
    defer container.Terminate(ctx)
    
    // 获取连接信息并运行测试
    host, _ := container.Host(ctx)
    port, _ := container.MappedPort(ctx, "5432")
    
    // 测试数据库集成
}
```

## 部署模式

### 单体应用
```go
func main() {
    // 初始化组件
    llm, _ := openai.New()
    memory := memory.NewBuffer()
    chain := chains.NewConversationChain(llm, memory)
    
    // 启动 HTTP 服务器
    http.HandleFunc("/chat", func(w http.ResponseWriter, r *http.Request) {
        response, _ := chain.Run(r.Context(), r.FormValue("message"))
        w.Write([]byte(response))
    })
    
    log.Fatal(http.ListenAndServe(":8080", nil))
}
```

### 微服务架构
```go
// 聊天服务
type ChatService struct {
    llm    llms.Model
    memory memory.Memory
}

func (s *ChatService) HandleChat(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
    chain := chains.NewConversationChain(s.llm, s.memory)
    result, err := chain.Run(ctx, req.Message)
    if err != nil {
        return nil, err
    }
    
    return &ChatResponse{Message: result}, nil
}
```

### 云原生部署
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: langchaingo-app
spec:
  replicas: 3
  selector:
    matchLabels:
      app: langchaingo-app
  template:
    metadata:
      labels:
        app: langchaingo-app
    spec:
      containers:
      - name: app
        image: langchaingo-app:latest
        env:
        - name: OPENAI_API_KEY
          valueFrom:
            secretKeyRef:
              name: api-keys
              key: openai-key
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
```

这种架构为使用 LangChainGo 构建强大、可扩展的 AI 应用程序提供了坚实的基础，同时保持了 Go 语言的简洁性和性能特征。