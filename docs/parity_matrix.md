# LangChainGo 功能特性

本页面列出了 LangChainGo (一个用 Golang 构建的自然语言处理库) 的功能及其状态指示。我们目前的主要目标是与 Python 版本的 LangChain 功能保持一致。

请注意，本页面列出的是截至 2025 年 1 月 LangChainGo 项目的当前状态，部分功能可能仍在开发中或计划在未来版本中发布。Python LangChain 生态系统已经通过 LangChain 表达式语言 (LCEL)、用于复杂工作流的 LangGraph 以及模块化架构的改进而显著发展。如果您有兴趣为特定的集成或功能做出贡献，请随时从尚未完成的功能列表中选择，并告知我们，以便我们帮助您开始。

## 提示模板 (Prompt Templates)
| 功能特性                            | 状态 |
|------------------------------------|--------|
| 提示模板 (Prompt Templates)        | ✅     |
| 少样本提示模板 (Few Shot Prompt Template) | ✅     |
| 输出解析器 (Output Parsers)          | ✅     |
| 示例选择器 (Example Selectors)     | ✅     |


## 文本分割器 (Text Splitters)

| 功能特性                           | 状态 |
| --------------------------------- | ------ |
| 字符文本分割器 (Character Text Splitter) | ✅     |
| 递归字符文本分割器 (Recursive Character Text Splitter) | ✅     |
| Markdown 文本分割器 (Markdown Text Splitter) | ✅     |

## 链 (Chains)

| 功能特性                                | 状态 |
| -------------------------------------- | ------ |
| LLM 链 (LLM Chain)                     | ✅     |
| Stuff 合并文档链 (Stuff Combine Documents Chain) | ✅     |
| 检索问答链 (Retrieval QA Chain)          | ✅     |
| Map Reduce 合并文档链 (Map Reduce Combine Documents Chain) | ✅     |
| Refine 合并文档链 (Refine Combine Documents Chain) | ✅     |
| Map Rerank 合并文档链 (Map Rerank Combine Documents Chain) | ✅     |
| 聊天向量数据库链 (Chat Vector DB Chain)    | ❌     |
| 向量数据库问答链 (Vector DB QA Chain)    | ❌     |
| 分析文档链 (Analyze Document Chain)      | ❌     |
| 问答链 (Question Answering Chains)     | ✅     |
| 摘要链 (Summarization Chains)          | ✅     |
| 带来源的问答链 (Question Answering With Sources Chains) | ❌     |
| SQL 数据库链 (SQL Database Chain)        | ✅     |
| API 链 (API Chain)                     | ✅     |
| 转换链 (Transformation Chain)          | ✅     |
| 宪法链 (Constitutional Chain)          | ✅     |
| 对话链 (Conversational Chain)          | ✅     |
| 图问答链 (Graph QA Chain)                | ❌     |
| HyDE 链 (HyDE Chain)                   | ❌     |
| LLM Bash 链 (LLM Bash Chain)           | ❌     |
| LLM 数学链 (LLM Math Chain)            | ✅     |
| PAL 链 (PAL Chain)                     | ❌     |
| LLM 请求链 (LLM Requests Chain)        | ❌     |
| 审查链 (Moderation Chain)              | ❌     |
| 顺序链 (Sequential Chain)              | ✅     |
| 简单顺序链 (Simple Sequential Chain)   | ✅     |

## 代理 (Agents)

| 功能特性                             | 状态 |
| ----------------------------------- | ------ |
| zero-shot-react-description         | ✅     |
| chat-zero-shot-react-description    | ❌     |
| self-ask-with-search                | ❌     |
| react-docstore                      | ❌     |
| conversational-react-description    | ✅     |
| chat-conversational-react-description | ❌   |

## 记忆 (Memory)

| 功能特性                     | 状态 |
| ----------------------------| ------ |
| 缓冲区记忆 (Buffer Memory)        | ✅     |
| 缓冲区窗口记忆 (Buffer Window Memory) | ✅     |
| 摘要记忆 (Summary Memory)         | ❌     |
| 实体记忆 (Entity Memory)          | ❌     |
| 摘要缓冲区记忆 (Summary Buffer Memory) | ❌     |
| 知识图谱记忆 (Knowledge Graph Memory) | ❌     |

## 输出解析器 (Output Parsers)

| 功能特性                      | 状态 |
| ----------------------------| ------ |
| 简单解析器 (Simple)             | ✅     |
| 结构化解析器 (Structured)       | ✅     |
| 布尔值解析器 (Boolean)          | ✅     |
| 组合解析器 (Combining Parsers)  | ✅     |
| 正则表达式解析器 (Regex)        | ✅     |
| 正则表达式字典解析器 (Regex Dictionary) | ✅     |
| 逗号分隔列表解析器 (Comma separated list) | ✅     |
| 解析器修复器 (Parser fixer)     | ❌     |
| 重试解析器 (Retry)              | ❌     |
| 防护栏解析器 (Guardrail)        | ❌     |

## 文档加载器 (Document Loaders)

| 功能特性                         | 状态 |
|---------------------------------|--------|
| AssemblyAI                      | ✅      |
| Blob 加载器 (Blob Loaders)        | ❌      |
| Airbyte JSON                    | ❌      |
| Apify 数据集 (Apify Dataset)      | ❌      |
| Arxiv 文档加载器 (Arxiv Document Loader) | ❌      |
| Azlyrics                        | ❌      |
| Azure Blob 存储 (Azure Blob Storage) | ❌      |
| BigQuery                        | ❌      |
| Bilibili                        | ❌      |
| Blackboard                      | ❌      |
| 区块链 (Blockchain)             | ❌      |
| ChatGPT                         | ❌      |
| College Confidential            | ❌      |
| Confluence                      | ❌      |
| CoNLL-U                         | ❌      |
| CSV 加载器 (CSV Loader)           | ✅      |
| 数据帧 (Dataframe)                | ❌      |
| Diffbot                         | ❌      |
| 目录 (Directory)                  | ❌      |
| Discord                         | ❌      |
| DuckDB 加载器 (DuckDB Loader)     | ❌      |
| 电子邮件 (Email)                  | ❌      |
| EPUB                            | ❌      |
| Evernote                        | ❌      |
| Facebook 聊天 (Facebook Chat)   | ❌      |
| Figma                           | ❌      |
| GCS 目录 (GCS Directory)          | ❌      |
| GCS 文件 (GCS File)             | ❌      |
| Git                             | ❌      |
| Gitbook                         | ❌      |
| Google Drive                    | ❌      |
| 古腾堡图书 (Gutenberg Books)      | ❌      |
| HN (Hacker News)                | ❌      |
| HTML                            | ✅      |
| HTML BS (Beautiful Soup)        | ❌      |
| Hugging Face 数据集 (Hugging Face Dataset) | ❌      |
| iFixit                          | ❌      |
| 图像 (Image)                      | ❌      |
| 图像描述 (Image Captions)       | ❌      |
| IMSDb                           | ❌      |
| Markdown                        | ✅      |
| MediaWiki XML                   | ❌      |
| Modern Treasury                 | ❌      |
| Notebook (Jupyter)              | ❌      |
| Notion                          | ✅      |
| Notion 数据库 (Notion Database)   | ❌      |
| Obsidian                        | ❌      |
| OneDrive                        | ❌      |
| OneDrive 文件 (OneDrive File)   | ❌      |
| PDF                             | ✅      |
| PowerPoint                      | ❌      |
| Python                          | ❌      |
| ReadTheDocs                     | ❌      |
| Reddit                          | ❌      |
| Roam                            | ❌      |
| RTF                             | ❌      |
| S3 目录 (S3 Directory)          | ❌      |
| S3 文件 (S3 File)               | ❌      |
| Sitemap                         | ❌      |
| Slack 目录 (Slack Directory)    | ❌      |
| Spreedly                        | ❌      |
| SRT (字幕格式)                    | ❌      |
| Stripe                          | ❌      |
| Telegram                        | ❌      |
| 文本 (Text)                       | ✅      |
| TOML                            | ❌      |
| Twitter                         | ❌      |
| Unstructured (非结构化数据)       | ❌      |
| URL                             | ❌      |
| URL Playwright                  | ❌      |
| URL Selenium                    | ❌      |
| Web Base (网页基础加载器)         | ❌      |
| WhatsApp                        | ❌      |
| Word 文档 (Word Document)       | ❌      |
| YouTube                         | ❌      |

## 现代 LangChain 功能 (Python v0.3+)

这些是 LangChain Python 中较新的功能，代表了该框架当前的发展方向：

| 功能特性                            | Go 状态 | 说明 |
|------------------------------------|-----------|-------|
| LangChain 表达式语言 (LCEL)       | ❌      | 声明式链组合与优化执行 |
| LangGraph 集成                     | ❌      | 具有状态管理和分支的复杂工作流 |
| 异步/流式支持 (Async/Streaming Support) | ✅      | 部分 - 提供基本异步支持 |
| 模块化架构 (langchain-core)       | ❌  | 分离的核心抽象和集成 |
| 生产级记忆管理 (Production Memory Management) | ❌      | ChatMessageHistory 和长期记忆 |
| 增强型代理框架 (Enhanced Agent Framework) | ❌      | 工具调用和多代理编排 |
| 向量存储 RAG 模式 (Vector Store RAG Patterns) | ✅      | 部分 - 支持基本 RAG |
| 结构化输出解析 (Structured Output Parsing) | ✅      | 提供 JSON、XML、YAML 解析 |

## 向量存储 (Vector Stores)

| 功能特性                            | 状态 |
|------------------------------------|--------|
| AlloyDB (PostgreSQL + pgvector)    | ✅     |
| Azure AI 搜索 (Azure AI Search)    | ✅     |
| AWS Bedrock 知识库 (AWS Bedrock Knowledge Bases) | ✅     |
| Chroma                             | ✅     |
| Cloud SQL (PostgreSQL + pgvector)  | ✅     |
| Milvus                             | ✅     |
| MongoDB Atlas 向量搜索 (MongoDB Atlas Vector Search) | ✅     |
| OpenSearch                         | ✅     |
| PGVector (PostgreSQL)              | ✅     |
| Pinecone                           | ✅     |
| Qdrant                             | ✅     |
| Redis 向量 (Redis Vector)          | ✅     |
| Weaviate                           | ✅     |
| FAISS                              | ❌     |
| Elastic Search                     | ❌     |

## LLM 提供商 (LLM Providers)

| 功能特性                            | 状态 |
|------------------------------------|--------|
| OpenAI                             | ✅     |
| Anthropic (Claude)                 | ✅     |
| Google AI (Gemini)                 | ✅     |
| Google Vertex AI                   | ✅     |
| AWS Bedrock                        | ✅     |
| Mistral AI                         | ✅     |
| Cohere                             | ✅     |
| Hugging Face                       | ✅     |
| Ollama                             | ✅     |
| LlamaFile                          | ✅     |
| 本地 LLM (Local LLM)               | ✅     |
| Groq                               | ✅     |
| WatsonX                            | ✅     |
| Ernie (百度文心)                   | ✅     |
| Cloudflare Workers AI              | ✅     |
| Maritaca AI                        | ✅     |
| NVIDIA                             | ✅     |
| Perplexity                         | ✅     |
| DeepSeek                           | ✅     |
| 伪 LLM (Fake LLM, 用于测试)        | ✅     |

## 嵌入 (Embeddings)

| 功能特性                            | 状态 |
|------------------------------------|--------|
| OpenAI                             | ✅     |
| AWS Bedrock                        | ✅     |
| Google Vertex AI                   | ✅     |
| Hugging Face                       | ✅     |
| Cybertron (本地)                   | ✅     |
| Jina AI                            | ✅     |
| VoyageAI                           | ✅     |
| Mistral                            | ✅     |
| Cohere                             | ❌     |
| Azure OpenAI                       | ❌     |

## 工具 (Tools)

| 功能特性                            | 状态 |
|------------------------------------|--------|
| 计算器 (Calculator)                 | ✅     |
| DuckDuckGo 搜索 (DuckDuckGo Search) | ✅     |
| SerpAPI                            | ✅     |
| 维基百科 (Wikipedia)                | ✅     |
| 网页抓取器 (Web Scraper)            | ✅     |
| SQL 数据库 (SQL Database)           | ✅     |
| Zapier                             | ✅     |
| Metaphor 搜索 (Metaphor Search)    | ✅     |
| Perplexity 搜索 (Perplexity Search)| ✅     |
| Python REPL                        | ❌     |
| Bash                               | ❌     |
| 文件系统 (File System)              | ❌     |
| 人工工具 (Human Tool)               | ❌     |

## 回调 (Callbacks)

| 功能特性                            | 状态 |
|------------------------------------|--------|
| 基本回调 (Basic Callbacks)          | ✅     |
| 流式回调 (Streaming Callbacks)      | ✅     |
| 日志回调 (Log Callbacks)            | ✅     |
| 组合回调 (Combining Callbacks)      | ✅     |
| 代理最终流 (Agent Final Stream)     | ✅     |
| 自定义回调 (Custom Callbacks)       | ✅     |
| LangSmith 集成                     | ❌     |
| Wandb 集成                         | ❌     |
| MLflow 集成                        | ❌     |

## 聊天消息历史 (Chat Message History)

| 功能特性                            | 状态 |
|------------------------------------|--------|
| 内存历史 (In-Memory History)        | ✅     |
| SQLite3 历史 (SQLite3 History)    | ✅     |
| MongoDB 历史 (MongoDB History)      | ✅     |
| AlloyDB 历史 (AlloyDB History)    | ✅     |
| Cloud SQL 历史 (Cloud SQL History)| ✅     |
| Zep 记忆 (Zep Memory)               | ✅     |
| Redis 历史 (Redis History)        | ❌     |
| DynamoDB 历史 (DynamoDB History)  | ❌     |
| Cassandra 历史 (Cassandra History)| ❌     |

## 检索器 (Retrievers)

| 功能特性                            | 状态 |
|------------------------------------|--------|
| 向量存储检索器 (Vector Store Retriever) | ✅     |
| 上下文压缩 (Contextual Compression) | ❌     |
| 多查询检索器 (Multi Query Retriever)| ❌     |
| 父文档检索器 (Parent Document Retriever) | ❌     |
| 自查询检索器 (Self Query Retriever)| ❌     |
| 时间加权检索器 (Time Weighted Retriever) | ❌     |
| 集成检索器 (Ensemble Retriever)     | ❌     |

## 实验性功能 (Experimental Features)

| 功能特性                            | 状态 |
|------------------------------------|--------|
| 实验性包 (Experimental Package)     | ✅     |
| 社区贡献 (Community Contributions)| ✅     |

## 附加组件 (Additional Components)

| 功能特性                            | 状态 |
|------------------------------------|--------|
| JSON Schema 支持                   | ✅     |
| HTTP 工具 (HTTP Utilities)        | ✅     |
| 图像工具 (Image Utilities)        | ✅     |
| 缓存 (LLM 响应) (Caching (LLM responses)) | ✅     |
| Token 计数 (Token Counting)       | ✅     |
| 异步支持 (Async Support)            | ✅     |
| 流式支持 (Streaming Support)        | ✅     |
