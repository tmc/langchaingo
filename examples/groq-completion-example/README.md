# Groq Completion Example with LangChain Go

Hello there! 👋 This example demonstrates how to use the Groq API with LangChain Go to generate creative text completions. Let's break down what this exciting little program does!

## What This Example Does

1. **Environment Setup**: 
   - The program starts by loading environment variables from a `.env` file. This is where you'll store your Groq API key.

2. **Groq LLM Configuration**:
   - It sets up a Large Language Model (LLM) client using Groq's API, which is compatible with the OpenAI interface.
   - The model used is "llama3-8b-8192", a powerful language model hosted by Groq.

3. **Text Generation**:
   - The example prompts the model to "Write a long poem about how golang is a fantastic language."
   - It uses various parameters like temperature (0.8) and max tokens (4096) to control the output.

4. **Streaming Output**:
   - As the model generates text, it streams the output directly to the console, allowing you to see the poem being created in real-time!

## Cool Features

- **Real-time Streaming**: Watch as the AI crafts the poem word by word!
- **Customizable**: You can easily modify the prompt or adjust generation parameters.
- **Groq Integration**: Showcases how to use Groq's powerful models with LangChain Go.

## Running the Example

1. Make sure you have a Groq API key and set it in your `.env` file as `GROQ_API_KEY=your_api_key_here`.
2. Run the program and watch as it generates a creative poem about Golang right before your eyes!

Enjoy exploring the creative possibilities with Groq and LangChain Go! 🚀🐹
