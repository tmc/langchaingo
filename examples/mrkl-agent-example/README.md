# MRKL Agent Example 🤖🔍

Hello there! Welcome to this exciting example of a MRKL (Modular Reasoning, Knowledge, and Language) agent using the LangChain Go library. This nifty little program showcases how to create an AI-powered agent that can answer complex questions by combining different tools and language models. Let's dive in!

## What Does This Example Do? 🎭

This example demonstrates:

1. Setting up an OpenAI language model
2. Initializing a search tool (SerpAPI)
3. Creating a MRKL agent with multiple tools
4. Asking the agent a multi-step question

The agent is tasked with answering the following question:

> "Who is Olivia Wilde's boyfriend? What is his current age raised to the 0.23 power?"

To answer this, the agent needs to:
- Search for information about Olivia Wilde's boyfriend
- Find out his current age
- Use a calculator to compute the age raised to the 0.23 power

## How It Works 🛠️

1. The program sets up an OpenAI language model and a SerpAPI search tool.
2. It creates an agent with two tools: a calculator and the search tool.
3. The agent is initialized with a "zero-shot react description" approach, meaning it doesn't require specific examples to understand how to use the tools.
4. The question is passed to the agent, which then uses its tools and the language model to formulate an answer.
5. The answer is printed to the console.

## Running the Example 🏃‍♀️

To run this example, make sure you have your OpenAI API key and SerpAPI key set as environment variables. Then, simply execute the `main()` function, and watch the magic happen!

## Why This is Cool 😎

This example showcases the power of combining language models with external tools to solve complex, multi-step problems. It's a great demonstration of how AI can be used to augment human intelligence and automate research tasks.

So go ahead, give it a try, and see how the MRKL agent tackles this intriguing question about Olivia Wilde's boyfriend! 🎉🧠
