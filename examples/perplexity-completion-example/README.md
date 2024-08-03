# Perplexity Completion Example with LangChain Go

Hello there! 👋 This example demonstrates how to use the Perplexity API with LangChain Go to generate creative text completions. Let's break down what this exciting little program does!

## What This Example Does

1. **Environment Setup**:
   - The program starts by loading environment variables from a `.env` file. This is where you'll store your Perplexity API key.

2. **Perplexity LLM Configuration**:
   - It sets up a Large Language Model (LLM) client using Perplexity's API, which is compatible with the OpenAI interface.
   - The model used is "llama-3.1-sonar-large-128k-online", a powerful language model hosted by Perplexity.

3. **Text Generation**:
   - The example prompts the model to "What is a prime number?"

4. **Streaming Output**:
   - As the model generates text, it streams the output directly to the console, allowing you to see the response being created in real-time!

## Cool Features

- **Real-time Streaming**: Watch as the AI crafts the response word by word!
- **Customizable**: You can easily modify the prompt or adjust generation parameters.
- **Perplexity Integration**: Showcases how to use Perplexity's powerful models with LangChain Go.

## Running the Example

1. Make sure you have a Perplexity API key and set it in your `.env` file as `PERPLEXITY_API_KEY=your_api_key_here`.
2. Run the program and watch as it generates a creative response about prime numbers right before your eyes!

You can read more about the perplexity API here: [Perplexity API Documentation](https://docs.perplexity.ai/docs/getting-started).

Enjoy exploring the creative possibilities with Perplexity and LangChain Go! 🚀🐹