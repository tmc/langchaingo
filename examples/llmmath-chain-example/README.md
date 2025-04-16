# LLM Math Chain Example

This example demonstrates how to use the LLM Math Chain from the `langchaingo` library to perform mathematical operations using natural language input.

## What it does

The program does the following:

1. Sets up an OpenAI language model (LLM) client.
2. Creates an LLM Math Chain using the OpenAI model.
3. Runs the chain with a mathematical question in natural language.
4. Prints the result of the calculation.

## How it works

The main functionality is in the `run()` function:

1. It initializes an OpenAI LLM client.
2. Creates an LLM Math Chain using the `chains.NewLLMMathChain()` function.
3. Runs the chain with the question "What is 1024 plus six times 9?" using `chains.Run()`.
4. Prints the output, which should be the calculated result.

## Running the example

To run this example, make sure you have set up your OpenAI API credentials properly. Then, execute the program, and it will output the result of the mathematical operation.

This example showcases how the LLM Math Chain can interpret natural language mathematical questions and provide accurate answers, combining the power of language models with mathematical computation.
