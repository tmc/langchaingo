# LangChainGo Features

This page provides a list of features and their status indicators for LangChainGo, a natural language processing library built in Golang. Our main focus at the moment is to reach parity with the Python version of LangChain.

Please note that this page lists the current state of the LangChainGo project as of January 2025, and some features may still be under development or planned for future releases. The Python LangChain ecosystem has evolved significantly with LangChain Expression Language (LCEL), LangGraph for complex workflows, and modular architecture improvements. If you are interested in contributing to a specific integration or feature, feel free to choose from the list of features that are not yet done and let us know so we can help you get started.

## Prompt Templates
| Feature                            | Status |
|------------------------------------|--------|
| Prompt Templates                   | Supported |
| Few Shot Prompt Template           | Supported |
| Output Parsers                     | Supported |
| Example Selectors                  | Supported |


## Text Splitters

| Feature                           | Status |
| --------------------------------- | ------ |
| Character Text Splitter           | Supported |
| Recursive Character Text Splitter | Supported |
| Markdown Text Splitter            | Supported |

## Chains

| Feature                                | Status |
| -------------------------------------- | ------ |
| LLM Chain                              | Supported |
| Stuff Combine Documents Chain          | Supported |
| Retrieval QA Chain                     | Supported |
| Map Reduce Combine Documents Chain     | Supported |
| Refine Combine Documents Chain         | Supported |
| Map Rerank Combine Documents Chain     | Supported |
| Chat Vector DB Chain                   | Not supported |
| Vector DB QA Chain                     | Not supported |
| Analyze Document Chain                 | Not supported |
| Question Answering Chains              | Supported |
| Summarization Chains                   | Supported |
| Question Answering With Sources Chains | Not supported |
| SQL Database Chain                     | Supported |
| API Chain                              | Supported |
| Transformation Chain                   | Supported |
| Constitutional Chain                   | Supported |
| Conversational Chain                   | Supported |
| Graph QA Chain                         | Not supported |
| HyDE Chain                             | Not supported |
| LLM Bash Chain                         | Not supported |
| LLM Math Chain                         | Supported |
| PAL Chain                              | Not supported |
| LLM Requests Chain                     | Not supported |
| Moderation Chain                       | Not supported |
| Sequential Chain                       | Supported |
| Simple Sequential Chain                | Supported |

## Agents

| Feature                             | Status |
| ----------------------------------- | ------ |
| zero-shot-react-description         | Supported |
| chat-zero-shot-react-description    | Not supported |
| self-ask-with-search                | Not supported |
| react-docstore                      | Not supported |
| conversational-react-description    | Supported |
| chat-conversational-react-description | Not supported |

## Memory

| Feature                     | Status |
| ----------------------------| ------ |
| Buffer Memory               | Supported |
| Buffer Window Memory        | Supported |
| Summary Memory              | Not supported |
| Entity Memory               | Not supported |
| Summary Buffer Memory       | Not supported |
| Knowledge Graph Memory      | Not supported |

## Output Parsers

| Feature                      | Status |
| ----------------------------| ------ |
| Simple                      | Supported |
| Structured                  | Supported |
| Boolean                     | Supported |
| Combining Parsers           | Supported |
| Regex                       | Supported |
| Regex Dictionary            | Supported |
| Comma separated list        | Supported |
| Parser fixer                | Not supported |
| Retry                       | Not supported |
| Guardrail                   | Not supported |

## Document Loaders

| Feature                         | Status |
|---------------------------------|--------|
| AssemblyAI                      | Supported |
| Blob Loaders                    | Not supported |
| Airbyte JSON                    | Not supported |
| Apify Dataset                   | Not supported |
| Arxiv Document Loader           | Not supported |
| Azlyrics                        | Not supported |
| Azure Blob Storage              | Not supported |
| BigQuery                        | Not supported |
| Bilibili                        | Not supported |
| Blackboard                      | Not supported |
| Blockchain                      | Not supported |
| ChatGPT                         | Not supported |
| College Confidential            | Not supported |
| Confluence                      | Not supported |
| CoNLL-U                         | Not supported |
| CSV Loader                      | Supported |
| Dataframe                       | Not supported |
| Diffbot                         | Not supported |
| Directory                       | Not supported |
| Discord                         | Not supported |
| DuckDB Loader                   | Not supported |
| Email                           | Not supported |
| EPUB                            | Not supported |
| Evernote                        | Not supported |
| Facebook Chat                   | Not supported |
| Figma                           | Not supported |
| GCS Directory                   | Not supported |
| GCS File                        | Not supported |
| Git                             | Not supported |
| Gitbook                         | Not supported |
| Google Drive                    | Not supported |
| Gutenberg Books                 | Not supported |
| HN                              | Not supported |
| HTML                            | Supported |
| HTML BS                         | Not supported |
| Hugging Face Dataset            | Not supported |
| iFixit                          | Not supported |
| Image                           | Not supported |
| Image Captions                  | Not supported |
| IMSDb                           | Not supported |
| Markdown                        | Supported |
| MediaWiki XML                   | Not supported |
| Modern Treasury                 | Not supported |
| Notebook                        | Not supported |
| Notion                          | Supported |
| Notion Database                 | Not supported |
| Obsidian                        | Not supported |
| OneDrive                        | Not supported |
| OneDrive File                   | Not supported |
| PDF                             | Supported |
| PowerPoint                      | Not supported |
| Python                          | Not supported |
| ReadTheDocs                     | Not supported |
| Reddit                          | Not supported |
| Roam                            | Not supported |
| RTF                             | Not supported |
| S3 Directory                    | Not supported |
| S3 File                         | Not supported |
| Sitemap                         | Not supported |
| Slack Directory                 | Not supported |
| Spreedly                        | Not supported |
| SRT                             | Not supported |
| Stripe                          | Not supported |
| Telegram                        | Not supported |
| Text                            | Supported |
| TOML                            | Not supported |
| Twitter                         | Not supported |
| Unstructured                    | Not supported |
| URL                             | Not supported |
| URL Playwright                  | Not supported |
| URL Selenium                    | Not supported |
| Web Base                        | Not supported |
| WhatsApp                        | Not supported |
| Word Document                   | Not supported |
| YouTube                         | Not supported |

## Modern LangChain Features (Python v0.3+)

These are newer features from LangChain Python that represent the current direction of the framework:

| Feature                            | Go Status | Notes |
|------------------------------------|-----------|-------|
| LangChain Expression Language (LCEL) | Not supported | Declarative chain composition with optimized execution |
| LangGraph Integration              | Not supported | Complex workflows with state management and branching |
| Async/Streaming Support            | Supported | Partial - Basic async support available |
| Modular Architecture (langchain-core) | Not supported | Separated core abstractions and integrations |
| Production Memory Management       | Not supported | ChatMessageHistory and long-term memory |
| Enhanced Agent Framework           | Not supported | Tool calling and multi-agent orchestration |
| Vector Store RAG Patterns          | Supported | Partial - Basic RAG supported |
| Structured Output Parsing          | Supported | JSON, XML, YAML parsing available |

## Vector Stores

| Feature                            | Status |
|------------------------------------|--------|
| AlloyDB (PostgreSQL + pgvector)    | Supported |
| Azure AI Search                    | Supported |
| AWS Bedrock Knowledge Bases        | Supported |
| Chroma                             | Supported |
| Cloud SQL (PostgreSQL + pgvector)  | Supported |
| Milvus                             | Supported |
| MongoDB Atlas Vector Search        | Supported |
| OpenSearch                         | Supported |
| PGVector (PostgreSQL)              | Supported |
| Pinecone                           | Supported |
| Qdrant                             | Supported |
| Redis Vector                       | Supported |
| Weaviate                           | Supported |
| FAISS                              | Not supported |
| Elastic Search                     | Not supported |

## LLM Providers

| Feature                            | Status |
|------------------------------------|--------|
| OpenAI                             | Supported |
| Anthropic (Claude)                 | Supported |
| Google AI (Gemini)                 | Supported |
| Google Vertex AI                   | Supported |
| AWS Bedrock                        | Supported |
| Mistral AI                         | Supported |
| Cohere                             | Supported |
| Hugging Face                       | Supported |
| Ollama                             | Supported |
| LlamaFile                          | Supported |
| Local LLM                          | Supported |
| Groq                               | Supported |
| WatsonX                            | Supported |
| Ernie (Baidu)                      | Supported |
| Cloudflare Workers AI              | Supported |
| Maritaca AI                        | Supported |
| NVIDIA                             | Supported |
| Perplexity                         | Supported |
| DeepSeek                           | Supported |
| Fake LLM (for testing)             | Supported |

## Embeddings

| Feature                            | Status |
|------------------------------------|--------|
| OpenAI                             | Supported |
| AWS Bedrock                        | Supported |
| Google Vertex AI                   | Supported |
| Hugging Face                       | Supported |
| Cybertron (local)                  | Supported |
| Jina AI                            | Supported |
| VoyageAI                           | Supported |
| Mistral                            | Supported |
| Cohere                             | Not supported |
| Azure OpenAI                       | Not supported |

## Tools

| Feature                            | Status |
|------------------------------------|--------|
| Calculator                         | Supported |
| DuckDuckGo Search                  | Supported |
| SerpAPI                            | Supported |
| Wikipedia                          | Supported |
| Web Scraper                        | Supported |
| SQL Database                       | Supported |
| Zapier                             | Supported |
| Metaphor Search                    | Supported |
| Perplexity Search                  | Supported |
| Python REPL                        | Not supported |
| Bash                               | Not supported |
| File System                        | Not supported |
| Human Tool                         | Not supported |

## Callbacks

| Feature                            | Status |
|------------------------------------|--------|
| Basic Callbacks                    | Supported |
| Streaming Callbacks                | Supported |
| Log Callbacks                      | Supported |
| Combining Callbacks                | Supported |
| Agent Final Stream                 | Supported |
| Custom Callbacks                   | Supported |
| LangSmith Integration              | Not supported |
| Wandb Integration                  | Not supported |
| MLflow Integration                 | Not supported |

## Chat Message History

| Feature                            | Status |
|------------------------------------|--------|
| In-Memory History                  | Supported |
| SQLite3 History                    | Supported |
| MongoDB History                    | Supported |
| AlloyDB History                    | Supported |
| Cloud SQL History                  | Supported |
| Zep Memory                         | Supported |
| Redis History                      | Not supported |
| DynamoDB History                   | Not supported |
| Cassandra History                  | Not supported |

## Retrievers

| Feature                            | Status |
|------------------------------------|--------|
| Vector Store Retriever             | Supported |
| Contextual Compression             | Not supported |
| Multi Query Retriever              | Not supported |
| Parent Document Retriever          | Not supported |
| Self Query Retriever               | Not supported |
| Time Weighted Retriever            | Not supported |
| Ensemble Retriever                 | Not supported |

## Experimental Features

| Feature                            | Status |
|------------------------------------|--------|
| Experimental Package               | Supported |
| Community Contributions            | Supported |

## Additional Components

| Feature                            | Status |
|------------------------------------|--------|
| JSON Schema Support                | Supported |
| HTTP Utilities                     | Supported |
| Image Utilities                    | Supported |
| Caching (LLM responses)            | Supported |
| Token Counting                     | Supported |
| Async Support                      | Supported |
| Streaming Support                  | Supported |
