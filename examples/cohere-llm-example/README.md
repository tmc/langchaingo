# Cohere Completion Example

Hello there! 👋 This example demonstrates how to use the Cohere language model for text completion using the LangChain Go library. Let's break down what this exciting little program does!

## What Does This Example Do?

1. **Sets Up the Cohere LLM**: The program initializes a Cohere language model using the `cohere.New()` function.

2. **Prepares the Input**: It defines a simple input prompt: "The first man to walk on the moon".

3. **Generates Completion**: Using the `llms.GenerateFromSinglePrompt()` function, it sends the input to the Cohere model and receives a completion.

4. **Displays the Result**: The generated completion is printed to the console.

5. **Token Counting**: As a bonus, it counts the number of tokens in both the input and output, giving you an idea of the model's verbosity.

## How to Run

1. Make sure you have Go installed on your system.
2. Set up your Cohere API key as an environment variable (the exact name depends on the LangChain Go implementation).
3. Run the program with `go run cohere_completion_example.go`.

## What to Expect

When you run this program, you'll see:
1. The generated completion based on the input prompt about the first man on the moon.
2. A token count in the format "input tokens / output tokens".

This example is perfect for anyone looking to get started with using Cohere's language model in their Go projects. It's a simple yet powerful demonstration of AI-powered text generation!

Happy coding! 🚀🌙
