# Welcome to LangChainGo

![gopher](https://pkg.go.dev/static/shared/icon/favicon.ico)

LangChainGo is the [Go Programming Language](https://go.dev/) port/fork of
[LangChain](https://www.langchain.com/).

LangChain is a framework for developing applications powered by language models. We believe that the most powerful and differentiated applications will not only call out to a language model via an API, but will also:

- _Be data-aware_: connect a language model to other sources of data
- _Be agentic_: allow a language model to interact with its environment

The LangChain framework is designed with the above principles in mind.

## Documentation Structure

_**Note**: These docs are for [LangChainGo](https://github.com/tmc/langchaingo). For documentation on [the Python version](https://github.com/langchain-ai/langchain), [head here](https://python.langchain.com/docs)._

Our documentation follows a structured approach to help you learn and use LangChain Go effectively:

### 📚 [Tutorials](./tutorials/)
Step-by-step guides to build complete applications. Perfect for learning LangChain Go from the ground up.

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
Deep explanations of LangChain Go's architecture and design principles.

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

[Here](https://pkg.go.dev/github.com/tmc/langchaingo) you can find the API reference for all of the modules in LangChain, as well as full documentation for all exported classes and functions.

## Additional Resources

Additional collection of resources we think may be useful as you develop your application!

- [LangChainHub](https://github.com/hwchase17/langchain-hub): The LangChainHub is a place to share and explore other prompts, chains, and agents.

- [Discord](https://discord.gg/6adMQxSpJS): Join us on our Discord to discuss all things LangChain!

- [Production Support](https://forms.gle/57d8AmXBYp8PP8tZA): As you move your LangChains into production, we'd love to offer more comprehensive support. Please fill out this form and we'll set up a dedicated support Slack channel.
