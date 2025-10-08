# Welcome to LangChainGo


LangChainGo is the [Go Programming Language](https://go.dev/) port/fork of
[LangChain](https://www.langchain.com/).

LangChain is a framework for developing applications powered by language models. We believe that the most powerful and differentiated applications will not only call out to a language model via an API, but will also:

- _Be data-aware_: connect a language model to other sources of data
- _Be agentic_: allow a language model to interact with its environment

The LangChain framework is designed with the above principles in mind.

## Documentation Structure

_**Note**: These docs are for [LangChainGo](https://github.com/vendasta/langchaingo). For documentation on [the Python version](https://github.com/langchain-ai/langchain), [head here](https://python.langchain.com/docs)._

Our documentation follows a structured approach to help you learn and use LangChainGo effectively:

### 📚 [Tutorials](./tutorials/)
Step-by-step guides to build complete applications. Perfect for learning LangChainGo from the ground up.

- **Getting Started**: [Quick setup with Ollama](./getting-started/guide-ollama.mdx) • [Quick setup with OpenAI](./getting-started/guide-openai.mdx)
- **Basic Applications**: Simple chat apps, Q&A systems, document summarization
- **Advanced Applications**: RAG systems, agents with tools, multi-modal apps
- **Production**: Deployment, optimization, monitoring

### 🛠️ [How-to Guides](./how-to/)
Practical solutions for specific problems. Find answers to "How do I...?" questions.

- **LLM Integration**: Configure providers, handle rate limits, implement streaming
- **Document Processing**: Load documents, implement search, optimize retrieval
- **Agent Development**: Create custom tools, multi-step reasoning, error handling
- **Production**: Project structure, logging, deployment, scaling

### 🧠 [Concepts](./concepts/)
Deep explanations of LangChainGo's architecture and design principles.

- **Core Architecture**: Framework design, interfaces, Go-specific patterns
- **Language Models**: Model abstraction, communication patterns, optimization
- **Agents & Memory**: Agent patterns, memory management, state persistence
- **Production**: Performance, reliability, security considerations

### 🔧 Components
Technical reference for all LangChainGo modules and their capabilities.

- **Model I/O**: LLMs, Chat Models, Embeddings, and Prompts
- **Data Connection**: Document loaders, vector stores, text splitters, retrievers
- **[Chains](./modules/chains/)**: Sequences of calls and end-to-end applications
- **[Memory](./modules/memory/)**: State persistence and conversation management
- **[Agents](./modules/agents/)**: Decision-making and autonomous behavior

## API Reference

[Here](https://pkg.go.dev/github.com/vendasta/langchaingo) you can find the API reference for all of the modules in LangChain, as well as full documentation for all exported classes and functions.

## Get Involved

- **[Contributing Guide](/docs/contributing)**: Learn how to contribute code and documentation
- **[GitHub Discussions](https://github.com/tmc/langchaingo/discussions)**: Join the conversation about LangChainGo
- **[GitHub Issues](https://github.com/tmc/langchaingo/issues)**: Report bugs or request features
