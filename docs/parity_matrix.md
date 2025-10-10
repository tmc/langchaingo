# LangChainGo Features

This page provides a list of features and their status indicators for LangChainGo, a natural language processing library built in Golang. Our main focus at the moment is to reach parity with the Python version of LangChain.

Please note that this page lists the current state of the LangChainGo project as of January 2025, and some features may still be under development or planned for future releases. The Python LangChain ecosystem has evolved significantly with LangChain Expression Language (LCEL), LangGraph for complex workflows, and modular architecture improvements. If you are interested in contributing to a specific integration or feature, feel free to choose from the list of features that are not yet done and let us know so we can help you get started.

## Prompt Templates
| Feature                            | Status |
|------------------------------------|--------|
| Prompt Templates                   | ✅     |
| Few Shot Prompt Template           | ✅     |
| Output Parsers                     | ✅     |
| Example Selectors                  | ✅     |


## Text Splitters

| Feature                           | Status |
| --------------------------------- | ------ |
| Character Text Splitter           | ✅     |
| Recursive Character Text Splitter | ✅     |
| Markdown Text Splitter            | ✅     |

## Chains

| Feature                                | Status |
| -------------------------------------- | ------ |
| LLM Chain                              | ✅     |
| Stuff Combine Documents Chain          | ✅     |
| Retrieval QA Chain                     | ✅     |
| Map Reduce Combine Documents Chain     | ✅     |
| Refine Combine Documents Chain         | ✅     |
| Map Rerank Combine Documents Chain     | ✅     |
| Chat Vector DB Chain                   | ❌     |
| Vector DB QA Chain                     | ❌     |
| Analyze Document Chain                 | ❌     |
| Question Answering Chains              | ✅     |
| Summarization Chains                   | ✅     |
| Question Answering With Sources Chains | ❌     |
| SQL Database Chain                     | ✅     |
| API Chain                              | ✅     |
| Transformation Chain                   | ✅     |
| Constitutional Chain                   | ✅     |
| Conversational Chain                   | ✅     |
| Graph QA Chain                         | ❌     |
| HyDE Chain                             | ❌     |
| LLM Bash Chain                         | ❌     |
| LLM Math Chain                         | ✅     |
| PAL Chain                              | ❌     |
| LLM Requests Chain                     | ❌     |
| Moderation Chain                       | ❌     |
| Sequential Chain                       | ✅     |
| Simple Sequential Chain                | ✅     |

## Agents

| Feature                             | Status |
| ----------------------------------- | ------ |
| zero-shot-react-description         | ✅     |
| chat-zero-shot-react-description    | ❌     |
| self-ask-with-search                | ❌     |
| react-docstore                      | ❌     |
| conversational-react-description    | ✅     |
| chat-conversational-react-description | ❌   |

## Memory

| Feature                     | Status |
| ----------------------------| ------ |
| Buffer Memory               | ✅     |
| Buffer Window Memory        | ✅     |
| Summary Memory              | ❌     |
| Entity Memory               | ❌     |
| Summary Buffer Memory       | ❌     |
| Knowledge Graph Memory      | ❌     |

## Output Parsers

| Feature                      | Status |
| ----------------------------| ------ |
| Simple                      | ✅     |
| Structured                  | ✅     |
| Boolean                     | ✅     |
| Combining Parsers           | ✅     |
| Regex                       | ✅     |
| Regex Dictionary            | ✅     |
| Comma separated list        | ✅     |
| Parser fixer                | ❌     |
| Retry                       | ❌     |
| Guardrail                   | ❌     |

## Document Loaders

| Feature                         | Status |
|---------------------------------|--------|
| AssemblyAI                      | ✅      |
| Blob Loaders                    | ❌      |
| Airbyte JSON                    | ❌      |
| Apify Dataset                   | ❌      |
| Arxiv Document Loader           | ❌      |
| Azlyrics                        | ❌      |
| Azure Blob Storage              | ❌      |
| BigQuery                        | ❌      |
| Bilibili                        | ❌      |
| Blackboard                      | ❌      |
| Blockchain                      | ❌      |
| ChatGPT                         | ❌      |
| College Confidential            | ❌      |
| Confluence                      | ❌      |
| CoNLL-U                         | ❌      |
| CSV Loader                      | ✅      |
| Dataframe                       | ❌      |
| Diffbot                         | ❌      |
| Directory                       | ❌      |
| Discord                         | ❌      |
| DuckDB Loader                   | ❌      |
| Email                           | ❌      |
| EPUB                            | ❌      |
| Evernote                        | ❌      |
| Facebook Chat                   | ❌      |
| Figma                           | ❌      |
| GCS Directory                   | ❌      |
| GCS File                        | ❌      |
| Git                             | ❌      |
| Gitbook                         | ❌      |
| Google Drive                    | ❌      |
| Gutenberg Books                 | ❌      |
| HN                              | ❌      |
| HTML                            | ✅      |
| HTML BS                         | ❌      |
| Hugging Face Dataset            | ❌      |
| iFixit                          | ❌      |
| Image                           | ❌      |
| Image Captions                  | ❌      |
| IMSDb                           | ❌      |
| Markdown                        | ✅      |
| MediaWiki XML                   | ❌      |
| Modern Treasury                 | ❌      |
| Notebook                        | ❌      |
| Notion                          | ✅      |
| Notion Database                 | ❌      |
| Obsidian                        | ❌      |
| OneDrive                        | ❌      |
| OneDrive File                   | ❌      |
| PDF                             | ✅      |
| PowerPoint                      | ❌      |
| Python                          | ❌      |
| ReadTheDocs                     | ❌      |
| Reddit                          | ❌      |
| Roam                            | ❌      |
| RTF                             | ❌      |
| S3 Directory                    | ❌      |
| S3 File                         | ❌      |
| Sitemap                         | ❌      |
| Slack Directory                 | ❌      |
| Spreedly                        | ❌      |
| SRT                             | ❌      |
| Stripe                          | ❌      |
| Telegram                        | ❌      |
| Text                            | ✅      |
| TOML                            | ❌      |
| Twitter                         | ❌      |
| Unstructured                    | ❌      |
| URL                             | ❌      |
| URL Playwright                  | ❌      |
| URL Selenium                    | ❌      |
| Web Base                        | ❌      |
| WhatsApp                        | ❌      |
| Word Document                   | ❌      |
| YouTube                         | ❌      |

## Modern LangChain Features (Python v0.3+)

These are newer features from LangChain Python that represent the current direction of the framework:

| Feature                            | Go Status | Notes |
|------------------------------------|-----------|-------|
| LangChain Expression Language (LCEL) | ❌      | Declarative chain composition with optimized execution |
| LangGraph Integration              | ❌      | Complex workflows with state management and branching |
| Async/Streaming Support            | ✅      | Partial - Basic async support available |
| Modular Architecture (langchain-core) | ❌  | Separated core abstractions and integrations |
| Production Memory Management       | ❌      | ChatMessageHistory and long-term memory |
| Enhanced Agent Framework           | ❌      | Tool calling and multi-agent orchestration |
| Vector Store RAG Patterns          | ✅      | Partial - Basic RAG supported |
| Structured Output Parsing          | ✅      | JSON, XML, YAML parsing available |

## Vector Stores

| Feature                            | Status |
|------------------------------------|--------|
| AlloyDB (PostgreSQL + pgvector)    | ✅     |
| Azure AI Search                    | ✅     |
| AWS Bedrock Knowledge Bases        | ✅     |
| Chroma                             | ✅     |
| Cloud SQL (PostgreSQL + pgvector)  | ✅     |
| Milvus                             | ✅     |
| MongoDB Atlas Vector Search        | ✅     |
| OpenSearch                         | ✅     |
| PGVector (PostgreSQL)              | ✅     |
| Pinecone                           | ✅     |
| Qdrant                             | ✅     |
| Redis Vector                       | ✅     |
| Weaviate                           | ✅     |
| FAISS                              | ❌     |
| Elastic Search                     | ❌     |

## LLM Providers

| Feature                            | Status |
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
| Local LLM                          | ✅     |
| Groq                               | ✅     |
| WatsonX                            | ✅     |
| Ernie (Baidu)                      | ✅     |
| Cloudflare Workers AI              | ✅     |
| Maritaca AI                        | ✅     |
| NVIDIA                             | ✅     |
| Perplexity                         | ✅     |
| DeepSeek                           | ✅     |
| Fake LLM (for testing)             | ✅     |

## Embeddings

| Feature                            | Status |
|------------------------------------|--------|
| OpenAI                             | ✅     |
| AWS Bedrock                        | ✅     |
| Google Vertex AI                   | ✅     |
| Hugging Face                       | ✅     |
| Cybertron (local)                  | ✅     |
| Jina AI                            | ✅     |
| VoyageAI                           | ✅     |
| Mistral                            | ✅     |
| Cohere                             | ❌     |
| Azure OpenAI                       | ❌     |

## Tools

| Feature                            | Status |
|------------------------------------|--------|
| Calculator                         | ✅     |
| DuckDuckGo Search                  | ✅     |
| SerpAPI                            | ✅     |
| Wikipedia                          | ✅     |
| Web Scraper                        | ✅     |
| SQL Database                       | ✅     |
| Zapier                             | ✅     |
| Metaphor Search                    | ✅     |
| Perplexity Search                  | ✅     |
| Python REPL                        | ❌     |
| Bash                               | ❌     |
| File System                        | ❌     |
| Human Tool                         | ❌     |

## Callbacks

| Feature                            | Status |
|------------------------------------|--------|
| Basic Callbacks                    | ✅     |
| Streaming Callbacks                | ✅     |
| Log Callbacks                      | ✅     |
| Combining Callbacks                | ✅     |
| Agent Final Stream                 | ✅     |
| Custom Callbacks                   | ✅     |
| LangSmith Integration              | ❌     |
| Wandb Integration                  | ❌     |
| MLflow Integration                 | ❌     |

## Chat Message History

| Feature                            | Status |
|------------------------------------|--------|
| In-Memory History                  | ✅     |
| SQLite3 History                    | ✅     |
| MongoDB History                    | ✅     |
| AlloyDB History                    | ✅     |
| Cloud SQL History                  | ✅     |
| Zep Memory                         | ✅     |
| Redis History                      | ❌     |
| DynamoDB History                   | ❌     |
| Cassandra History                  | ❌     |

## Retrievers

| Feature                            | Status |
|------------------------------------|--------|
| Vector Store Retriever             | ✅     |
| Contextual Compression             | ❌     |
| Multi Query Retriever              | ❌     |
| Parent Document Retriever          | ❌     |
| Self Query Retriever               | ❌     |
| Time Weighted Retriever            | ❌     |
| Ensemble Retriever                 | ❌     |

## Experimental Features

| Feature                            | Status |
|------------------------------------|--------|
| Experimental Package               | ✅     |
| Community Contributions            | ✅     |

## Additional Components

| Feature                            | Status |
|------------------------------------|--------|
| JSON Schema Support                | ✅     |
| HTTP Utilities                     | ✅     |
| Image Utilities                    | ✅     |
| Caching (LLM responses)            | ✅     |
| Token Counting                     | ✅     |
| Async Support                      | ✅     |
| Streaming Support                  | ✅     |
