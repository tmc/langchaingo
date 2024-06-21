# OpenAI Agent with LangChain Go Example

Welcome to this example of using OpenAI's GPT-4 model as an agent with the LangChain Go library! 🤖

## What does this example do?

This example demonstrates how to use the LangChain Go library to create an agent powered by OpenAI's GPT-4 model. The agent is equipped with a set of tools and can intelligently decide which tool to use based on the given input. Here's a breakdown of the main features:

1. **OpenAI Model Initialization**: The code sets up a connection to the GPT-4 Turbo model.

2. **Tool Definition**: A calculator tool is provided to the agent, enabling it to perform mathematical calculations.

3. **Agent Creation**: An OpenAI Functions Agent is created, combining the GPT-4 model with the defined tools.

4. **Executor Setup**: An executor is set up to run the agent, specifying the maximum number of iterations and returning intermediate steps.

5. **Agent Execution**: The agent is given an input query that involves both a mathematical calculation and a general question about Python.

## How it works

1. The program initializes the OpenAI model and sets up the agent with the calculator tool.
2. It creates an executor to run the agent, specifying the maximum number of iterations and to return intermediate steps.
3. The agent is given an input query: "what is 3 plus 3 and what is python".
4. The agent processes the input and intelligently decides which tool to use (in this case, the calculator for the math part).
5. The agent generates a response that includes the result of the calculation and a general explanation of Python.

## Why is this cool?

- **Intelligent Tool Selection**: The agent can dynamically select the appropriate tool based on the input query.
- **Multi-tasking**: The agent can handle multiple tasks within a single input, such as performing a calculation and answering a general question.
- **Extensibility**: Additional tools can be easily added to expand the agent's capabilities.

Try it out and see how the agent handles different input queries by leveraging the power of GPT-4 and the flexibility of LangChain Go! 🚀
