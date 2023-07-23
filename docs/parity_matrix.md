# LangChain Go Features

This page provides a list of features and their status indicators for LangChain Go, a natural language processing library built in Golang. Our main focus at the moment is to reach parity with the Python version of LangChain.

Please note that this page lists the current state of the LangChain Go project, and some features may still be under development or planned for future releases. If you are interested in contributing to a specific integration or feature, feel free to choose from the list of features that are not yet done and let us know so we can help you get started.

## Prompt Templates
| Feature                            | Status |
|------------------------------------|--------|
| Prompt Templates                   | ✅     |
| Few Shot Prompt Template           | ❌     |
| Output Parsers                     | ❌     |
| Example Selectors                  | ❌     |


## Text Splitters

| Feature                           | Status |
| --------------------------------- | ------ |
| Character Text Splitter           | ❌     |
| Recursive Character Text Splitter | ✅     |
| Markdown Text Splitter            | ❌     |

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
| Transformation Chain                   | ❌     |
| Constitutional Chain                   | ❌     |
| Conversational Chain                   | ❌     |
| Graph QA Chain                         | ❌     |
| HyDE Chain                             | ❌     |
| LLM Bash Chain                         | ❌     |
| LLM Math Chain                         | ✅     |
| PAL Chain                              | ❌     |
| LLM Requests Chain                     | ❌     |
| Moderation Chain                       | ❌     |
| Sequential Chain                       | ❌     |
| Simple Sequential Chain                | ❌     |

## Agents

| Feature                             | Status |
| ----------------------------------- | ------ |
| zero-shot-react-description         | ✅     |
| chat-zero-shot-react-description    | ❌     |
| self-ask-with-search                | ❌     |
| react-docstore                      | ❌     |
| conversational-react-description    | ❌     |
| chat-conversational-react-description | ❌   |

## Memory

| Feature                     | Status |
| ----------------------------| ------ |
| Buffer Memory               | ✅     |
| Buffer Window Memory        | ❌     |
| Summary Memory              | ❌     |
| Entity Memory               | ❌     |
| Summary Buffer Memory       | ❌     |
| Knowledge Graph Memory      | ❌     |

## Output Parsers

| Feature                      | Status |
| ----------------------------| ------ |
| Simple                      | ✅     |
| Structured                  | ✅     |
| Boolean                     | ❌     |
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
| Markdown                        | ❌      |
| MediaWiki XML                   | ❌      |
| Modern Treasury                 | ❌      |
| Notebook                        | ❌      |
| Notion                          | ❌      |
| Notion Database                 | ❌      |
| Obsidian                        | ❌      |
| OneDrive                        | ❌      |
| OneDrive File                   | ❌      |
| PDF                             | ❌      |
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
