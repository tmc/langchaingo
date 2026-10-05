---
title: "Getting Started: LLMChain"
draft: true
---

# Getting Started: LLMChain

<Info>
[Conceptual Guide](https://python.langchain.com/docs/modules/chains)
</Info>

An `LLMChain` is a simple chain that adds some functionality around language models. It is used widely throughout LangChain, including in other chains and agents.

An `LLMChain` consists of a `PromptTemplate` and a language model (either an [LLM](../model_io/models/llms/index.md) or [chat model](../model_io/models/chat/index.md)).

## Usage with LLMs

We can construct an LLMChain which takes user input, formats it with a PromptTemplate, and then passes the formatted response to an LLM:

<Code src="/examples/llm-chain-example/llm_chain.go" title="examples/llm-chain-example/llm_chain.go" link />
