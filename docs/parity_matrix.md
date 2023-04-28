# LangChain Go Features

This page provides a list of features and their status indicators for LangChain Go, a natural language processing library built in Golang. Our main focus at the moment is to reach parity with the Python version of LangChain. 

Please note that this page lists the current state of the LangChain Go project, and some features may still be under development or planned for future releases. If you are interested in contributing to a specific integration or feature, feel free to choose from the list of features that are not yet done and let us know so we can help you get started. 

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
| Map Reduce Combine Documents Chain     | ✅     |
| Refine Combine Documents Chain         | ✅     |
| Map Rerank Combine Documents Chain     | ❌ [#68](https://github.com/tmc/langchaingo/issues/68) |
| Chat Vector DB Chain                   | ✅     |
| Vector DB QA Chain                     | ✅     |
| Analyze Document Chain                 | ✅     |
| Question Answering Chains              | ✅     |
| Summarization Chains                   | ✅     |
| Question Answering With Sources Chains | ❌     |
| SQL Database Chain                     | ✅     |
| API Chain                              | ❌     |
| Transformation Chain                   | ❌     |
| Constitutional Chain                   | ❌     |
| Conversational Chain                   | ✅     |
| Graph QA Chain                         | ❌     |
| HyDE Chain                             | ❌     |
| LLM Bash Chain                         | ❌     |
| LLM Math Chain                         | ❌     |
| PAL Chain                              | ❌     |
| LLM Requests Chain                     | ❌     |
| Moderation Chain                       | ❌     |
| Sequential Chain                       | ❌     |
| Simple Sequential Chain                | ❌     |

## Agents

| Feature                             | Status |
| ----------------------------------- | ------ |
| zero-shot-react-description         | ✅     |
| chat-zero-shot-react-description    | ✅     |
| self-ask-with-search                | ❌     |
| react-docstore                      | ❌     |
| conversational-react-description    | ❌     |
| chat-conversational-react-description | ✅   |

## Memory

| Feature                      | Status |
| ----------------------------| ------ |
| Buffer Memory               | ✅     |
| Buffer Window Meory         | ✅     |
| Summary Memory              | ❌     |
| Entity Memory               | ❌     |
| Summary Buffer Memory       | ❌     |
| Knowledge Graph Memory      | ❌     |
