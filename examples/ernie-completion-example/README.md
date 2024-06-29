# ERNIE Completion Example

Hello there! 👋 This example demonstrates how to use the ERNIE language model with the LangChain Go library. Let's break down what this exciting code does!

## What This Example Does

This Go program showcases two main features:

1. **Text Generation**: It generates text using the ERNIE-Bot model.
2. **Text Embedding**: It creates embeddings for text input.

### Text Generation

The program initializes an ERNIE language model and generates text based on a prompt. Here's what it does:

- Sets up the ERNIE model (specifically ERNIE-Bot).
- Generates text from the prompt "介绍一下你自己" (which means "Introduce yourself" in Chinese).
- Uses a temperature of 0.8 for some creative variety in the output.
- Implements streaming, printing each chunk of the generated text as it's produced.

### Text Embedding

After text generation, the program demonstrates how to create embeddings:

- Initializes an embedder using the ERNIE model.
- Creates an embedding for the text "你好" (which means "Hello" in Chinese).
- Prints out the dimensions of the resulting embedding.

## Cool Features

- **Streaming Output**: The example shows how to stream the generated text, which is great for real-time applications!
- **Customizable**: While not used in this example, it mentions how you can customize the model with specific authentication info and model selection.
- **Multilingual**: The example uses Chinese prompts, showcasing ERNIE's multilingual capabilities.

## How to Use

To run this example, make sure you have the necessary dependencies installed and the appropriate ERNIE API credentials set up. Then, simply run the Go file!

Remember, for actual use, you'd want to include your API key and secret key using `ernie.WithAKSK(apiKey,secretKey)` when initializing the model.

Have fun exploring the capabilities of ERNIE with LangChain Go! 🚀🤖
