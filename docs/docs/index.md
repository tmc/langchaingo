# Welcome to LangChainGo

![gopher](https://pkg.go.dev/static/shared/icon/favicon.ico)

LangchainGo is the [Go Programming Language](https://go.dev/) port/fork of
[LangChain](https://www.langchain.com/).

LangChain is a framework for developing applications powered by language models. We believe that the most powerful and differentiated applications will not only call out to a language model via an API, but will also:

- _Be data-aware_: connect a language model to other sources of data
- _Be agentic_: allow a language model to interact with its environment

The LangChain framework is designed with the above principles in mind.

## Getting Started

_**Note**: These docs are for [LangChainGo](https://github.com/tmc/langchaingo). For documentation on [the Python version](https://github.com/langchain-ai/langchain), [head here](https://python.langchain.com/docs)._

Checkout the guide below for a walkthrough of how to get started using LangChain to create a Language Model application.

- [Quickstart, using Ollama](./getting-started/guide-ollama.mdx)
- [Quickstart, using OpenAI](./getting-started/guide-openai.mdx)

## Components

There are several main modules that LangChain provides support for. For each module we provide some examples to get started and get familiar with some of the concepts. 

These modules are, in increasing order of complexity:

- [Models](./modules/model_io/models/): This includes integrations with a variety of LLMs, Chat Models and Embeddings models.

- [Prompts](./modules/model_io/prompts/): This includes prompt Templates and functionality to work with prompts like Output Parsers and Example Selectors

<!-- - [Data connection](./modules/data_connection/): This includes patterns and functionality for working with your own data, and making it ready to interact with language models (including document loaders, vectorstores, text splitters and retrievers). -->

<!-- - [Chains](./modules/chains/): Chains go beyond just a single LLM call, and are sequences of calls (whether to an LLM or a different utility). LangChain provides a standard interface for chains, lots of integrations with other tools, and end-to-end chains for common applications.-->

<!-- - [Memory](./modules/memory/): Memory is the concept of persisting state between calls of a chain/agent. LangChain provides a standard interface for memory, a collection of memory implementations, and examples of chains/agents that use memory. -->

<!-- - [Agents](./modules/agents/): Agents involve an LLM making decisions about which Actions to take, taking that Action, seeing an Observation, and repeating that until done. LangChain provides a standard interface for agents, a selection of agents to choose from, and examples of end-to-end agents. -->

## API Reference

[Here](https://pkg.go.dev/github.com/tmc/langchaingo) you can find the API reference for all of the modules in LangChain, as well as full documentation for all exported classes and functions.

## Additional Resources

Additional collection of resources we think may be useful as you develop your application!

- [LangChainHub](https://github.com/hwchase17/langchain-hub): The LangChainHub is a place to share and explore other prompts, chains, and agents.

- [Discord](https://discord.gg/6adMQxSpJS): Join us on our Discord to discuss all things LangChain!

- [Production Support](https://forms.gle/57d8AmXBYp8PP8tZA): As you move your LangChains into production, we'd love to offer more comprehensive support. Please fill out this form and we'll set up a dedicated support Slack channel.
