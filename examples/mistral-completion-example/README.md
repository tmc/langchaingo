# Mistral Completion Example

Welcome to this fun and informative example of using the Mistral language model with LangChain in Go! 🚀🌙

This example demonstrates how to generate text completions using the Mistral AI model through the LangChain Go library. It showcases two different approaches: streaming and non-streaming completions.

## What This Example Does

1. **Setup**: The code initializes a Mistral language model using the `mistral.New()` function, specifying the "open-mistral-7b" model.

2. **Streaming Completion**:
   - It generates a completion for the question "Who was the first man to walk on the moon?"
   - The response is streamed in real-time, with each chunk printed to the console as it arrives.
   - This demonstrates how to handle streaming responses, which can be useful for displaying results progressively.

3. **Non-Streaming Completion**:
   - It generates a completion for the question "Who was the first man to go to space?"
   - This time, it waits for the full response before printing it.
   - It uses a different model ("mistral-small-latest") and a lower temperature setting for variety.

## Key Features

- **Model Selection**: Shows how to specify different Mistral models.
- **Temperature Control**: Demonstrates adjusting the randomness of outputs using the temperature parameter.
- **Streaming vs. Non-Streaming**: Illustrates both real-time and batch completion methods.

## Running the Example

To run this example, make sure you have the necessary dependencies installed and your Mistral API credentials set up. Then, simply execute the Go file:

```
go run mistral_completion_example.go
```

Enjoy exploring the capabilities of Mistral AI with LangChain in Go! 🎉👨‍🚀
